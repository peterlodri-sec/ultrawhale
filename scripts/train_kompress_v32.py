#!/usr/bin/env python3
"""Fine-tune Kompress v3.1 — higher must-keep weight + hard label override.

Changes from train_kompress.py (v3):
  - must_keep_weight: 3.0 -> 6.0
  - Hard label=1 override for any token matching _MUST_KEEP_RE, regardless
    of whether it appears in the compressed reference.
  - Default base model: PeetPedro/kompress-v3 (fine-tune from v3, not v2)


# Disable torch.compile (inductor needs a C compiler not present in most GPU images)
import torch._dynamo
torch._dynamo.config.suppress_errors = True

Architecture:
  ModernBERT-base (frozen) + Head1 (token classifier) + Head2 (span CNN)
  LoRA r=16 applied to last 4 encoder layers (q/k/v projections)

Training objective:
  Weighted BCE — label=1 (keep) for tokens in compressed reference,
  with higher weights for must-keep tokens (numbers, identifiers, paths).

Silver labels:
  Token gets label=1 if the word it belongs to appears in the compressed
  reference (free_response). Must-keep pattern tokens always get label=1.

Output:
  ./kompress-v3-finetuned/  (PyTorch checkpoint + tokenizer)

Usage (on vast.ai GPU):
  pip install -q transformers peft datasets torch accelerate
  python train_kompress.py \
      --data data/kompress_train.jsonl \
      --base-model chopratejas/kompress-v2-base \
      --output kompress-v3-finetuned \
      --epochs 3 \
      --batch-size 16
"""
from __future__ import annotations

import argparse
import json
import logging
import re
from pathlib import Path
from typing import Optional

import torch
import torch.nn as nn
from torch.utils.data import DataLoader, Dataset
from transformers import AutoModel, AutoTokenizer
try:
    from peft import get_peft_model, LoraConfig
    _PEFT_AVAILABLE = True
except ImportError:
    _PEFT_AVAILABLE = False

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

BASE_ENCODER = "answerdotai/ModernBERT-base"

_MUST_KEEP_RE = re.compile(
    r"\d+(\.\d+)?"           # numbers
    r"|[A-Z_]{2,}"            # ALLCAPS
    r"|[a-z_]+\.[a-z_]+"     # dotted.paths
    r"|/[a-z/._-]{2,}"       # unix paths
    r"|\.[a-z]{2,4}\b"       # extensions
    r"|--?[a-zA-Z][\w-]*"       # flags
    r"|\b[A-Z][a-z]+[A-Z]\w*"  # CamelCase
)


def _word_set(text: str) -> set[str]:
    """Lowercase word tokens for fuzzy alignment."""
    return set(re.findall(r"\b[a-z]{3,}\b", text.lower()))


def _silver_labels(
    token_strings: list[str],
    reference_words: set[str],
) -> tuple[list[int], list[float]]:
    """Per-token labels and weights from reference alignment."""
    labels: list[int] = []
    weights: list[float] = []
    for tok in token_strings:
        clean = re.sub(r"[^\w]", "", tok).lower()
        is_must = bool(_MUST_KEEP_RE.search(tok))
        in_ref = clean in reference_words or len(clean) < 3
        # Hard override: must-keep tokens always get label=1 regardless of reference
        label = 1 if (is_must or in_ref) else 0
        weight = 10.0 if is_must else (1.0 if label == 1 else 0.5)
        labels.append(label)
        weights.append(weight)
    return labels, weights


class KompressDataset(Dataset):
    def __init__(self, path: str, tokenizer, max_length: int = 512):
        self.tokenizer = tokenizer
        self.max_length = max_length
        self.items: list[dict] = []
        with open(path) as f:
            for line in f:
                d = json.loads(line.strip())
                if d.get("text") and d.get("reference"):
                    self.items.append(d)
        logger.info("Dataset: %d items", len(self.items))

    def __len__(self) -> int:
        return len(self.items)

    def __getitem__(self, idx: int) -> dict:
        item = self.items[idx]
        text = item["text"]
        ref_words = _word_set(item["reference"])

        enc = self.tokenizer(
            text,
            max_length=self.max_length,
            truncation=True,
            padding=False,
            return_offsets_mapping=True,
            return_tensors="pt",
        )
        input_ids = enc["input_ids"][0]
        attention_mask = enc["attention_mask"][0]
        offsets = enc["offset_mapping"][0].tolist()

        # Decode each token for label assignment
        token_strings = [
            text[s:e] if e > s else self.tokenizer.decode([input_ids[i]])
            for i, (s, e) in enumerate(offsets)
        ]
        raw_labels, raw_weights = _silver_labels(token_strings, ref_words)

        # Special tokens ([CLS], [SEP]) always keep
        for i in range(len(raw_labels)):
            if offsets[i] == (0, 0):  # special token
                raw_labels[i] = 1
                raw_weights[i] = 0.0  # don't penalize special tokens

        return {
            "input_ids": input_ids,
            "attention_mask": attention_mask,
            "labels": torch.tensor(raw_labels, dtype=torch.long),
            "weights": torch.tensor(raw_weights, dtype=torch.float),
        }


def collate_fn(batch: list[dict]) -> dict:
    max_len = max(b["input_ids"].shape[0] for b in batch)
    input_ids = torch.zeros(len(batch), max_len, dtype=torch.long)
    attention_mask = torch.zeros(len(batch), max_len, dtype=torch.long)
    labels = torch.zeros(len(batch), max_len, dtype=torch.long)
    weights = torch.zeros(len(batch), max_len, dtype=torch.float)
    for i, b in enumerate(batch):
        L = b["input_ids"].shape[0]
        input_ids[i, :L] = b["input_ids"]
        attention_mask[i, :L] = b["attention_mask"]
        labels[i, :L] = b["labels"]
        weights[i, :L] = b["weights"]
    return {
        "input_ids": input_ids,
        "attention_mask": attention_mask,
        "labels": labels,
        "weights": weights,
    }


class HeadroomCompressorModel(nn.Module):
    """Mirrors kompress_compressor._get_model_class() for training."""

    def __init__(self, base_model_name: str = BASE_ENCODER):
        super().__init__()
        self.encoder = AutoModel.from_pretrained(base_model_name, attn_implementation="eager")
        hidden_size = self.encoder.config.hidden_size  # 768
        self.token_dropout = nn.Dropout(0.1)
        self.token_head = nn.Linear(hidden_size, 2)
        self.span_conv = nn.Sequential(
            nn.Conv1d(hidden_size, 256, kernel_size=5, padding=2),
            nn.GELU(),
            nn.Conv1d(256, 1, kernel_size=3, padding=1),
            nn.Sigmoid(),
        )

    def forward(self, input_ids, attention_mask):
        hidden = self.encoder(input_ids, attention_mask=attention_mask).last_hidden_state
        # Head 1
        token_logits = self.token_head(self.token_dropout(hidden))  # [B, L, 2]
        # Head 2
        span_scores = self.span_conv(hidden.transpose(1, 2)).squeeze(1)  # [B, L]
        return token_logits, span_scores


def load_v2_weights(model: HeadroomCompressorModel, model_id: str) -> None:
    """Load merged.pt from a local directory or HF repo."""
    local = Path(model_id) / "merged.pt"
    if local.exists():
        path = str(local)
    else:
        from huggingface_hub import hf_hub_download
        path = hf_hub_download(repo_id=model_id, filename="merged.pt")
    ckpt = torch.load(path, map_location="cpu", weights_only=True)
    missing_enc, _ = model.encoder.load_state_dict(ckpt["encoder_state_dict"], strict=False)
    model.token_head.load_state_dict(ckpt["token_head_state_dict"])
    span_key = "span_conv_state_dict" if "span_conv_state_dict" in ckpt else "span_head_state_dict"
    model.span_conv.load_state_dict(ckpt[span_key])
    logger.info("Loaded v2 weights. Missing encoder keys: %d", len(missing_enc))


def apply_lora(model: HeadroomCompressorModel) -> HeadroomCompressorModel:
    """Freeze encoder, train heads only. Falls back gracefully from LoRA failures."""
    # Discover actual attention module names from the encoder
    attn_names: set[str] = set()
    for name, _ in model.encoder.named_modules():
        leaf = name.split(".")[-1]
        if leaf in ("Wqkv", "Wo", "query", "key", "value", "q_proj", "k_proj", "v_proj"):
            attn_names.add(leaf)

    if _PEFT_AVAILABLE and attn_names:
        try:
            config = LoraConfig(
                r=16,
                lora_alpha=32,
                target_modules=list(attn_names),
                lora_dropout=0.05,
                bias="none",
            )
            model.encoder = get_peft_model(model.encoder, config)
            model.encoder.print_trainable_parameters()
            logger.info("LoRA applied to: %s", attn_names)
            return model
        except Exception as e:
            logger.warning("LoRA failed (%s) — falling back to frozen encoder + heads only", e)

    # Fallback: freeze encoder, train only token_head + span_conv (~2M params)
    for param in model.encoder.parameters():
        param.requires_grad = False
    trainable = sum(p.numel() for p in model.parameters() if p.requires_grad)
    logger.info("Frozen encoder. Trainable params: %d (heads only)", trainable)
    return model


def train(
    data_path: str,
    base_model_id: str,
    output_dir: str,
    epochs: int = 5,
    batch_size: int = 16,
    lr: float = 2e-4,
    max_length: int = 512,
) -> None:
    device = (
        torch.device("cuda") if torch.cuda.is_available()
        else torch.device("mps") if torch.backends.mps.is_available()
        else torch.device("cpu")
    )
    logger.info("Device: %s", device)

    tokenizer = AutoTokenizer.from_pretrained(BASE_ENCODER)
    dataset = KompressDataset(data_path, tokenizer, max_length=max_length)
    loader = DataLoader(dataset, batch_size=batch_size, shuffle=True, collate_fn=collate_fn)

    model = HeadroomCompressorModel(BASE_ENCODER)
    try:
        load_v2_weights(model, base_model_id)
    except Exception as e:
        logger.warning("Could not load v2 weights (%s), training from scratch", e)

    model = apply_lora(model)
    model.to(device)

    # Only train LoRA + heads, keep frozen encoder backbone
    trainable = (
        list(model.encoder.parameters())  # LoRA params only (rest frozen by peft)
        + list(model.token_head.parameters())
        + list(model.span_conv.parameters())
    )
    optimizer = torch.optim.AdamW(trainable, lr=lr, weight_decay=0.01)
    total_steps = epochs * len(loader)
    scheduler = torch.optim.lr_scheduler.CosineAnnealingLR(optimizer, T_max=total_steps)

    token_criterion = nn.CrossEntropyLoss(reduction="none")

    for epoch in range(epochs):
        model.train()
        total_loss = 0.0
        for step, batch in enumerate(loader):
            ids = batch["input_ids"].to(device)
            mask = batch["attention_mask"].to(device)
            labels = batch["labels"].to(device)
            weights = batch["weights"].to(device)

            token_logits, span_scores = model(ids, mask)
            B, L, _ = token_logits.shape

            # Weighted token classification loss
            flat_logits = token_logits.view(B * L, 2)
            flat_labels = labels.view(B * L)
            flat_weights = weights.view(B * L)
            flat_mask = mask.view(B * L).float()

            loss_per_token = token_criterion(flat_logits, flat_labels)
            loss = (loss_per_token * flat_weights * flat_mask).sum() / (flat_mask.sum() + 1e-8)

            # Span consistency: span scores should correlate with token keep decisions
            token_probs = torch.softmax(token_logits, dim=-1)[:, :, 1]  # [B, L]
            span_target = (token_probs.detach() > 0.5).float()
            span_loss = nn.functional.binary_cross_entropy(
                span_scores * mask.float(), span_target * mask.float(), reduction="mean"
            )
            total = loss + 0.3 * span_loss

            optimizer.zero_grad()
            total.backward()
            nn.utils.clip_grad_norm_(trainable, 1.0)
            optimizer.step()
            scheduler.step()

            total_loss += total.item()
            if step % 20 == 0:
                logger.info(
                    "Epoch %d/%d step %d/%d loss=%.4f",
                    epoch + 1, epochs, step, len(loader), total.item()
                )

        logger.info("Epoch %d avg loss=%.4f", epoch + 1, total_loss / len(loader))

    # Merge LoRA back into encoder
    model.encoder = model.encoder.merge_and_unload()
    model.eval()

    # Save checkpoint
    out = Path(output_dir)
    out.mkdir(parents=True, exist_ok=True)
    tokenizer.save_pretrained(out)

    # Save in merged.pt format matching kompress-v2-base convention
    torch.save(
        {
            "encoder_state_dict": model.encoder.state_dict(),
            "token_head_state_dict": model.token_head.state_dict(),
            "span_head_state_dict": model.span_conv.state_dict(),
        },
        out / "merged.pt",
    )
    logger.info("Saved to %s", out)


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data/kompress_train.jsonl")
    ap.add_argument("--base-model", default="PeetPedro/kompress-v3")
    ap.add_argument("--output", default="kompress-v3-finetuned")
    ap.add_argument("--epochs", type=int, default=3)
    ap.add_argument("--batch-size", type=int, default=16)
    ap.add_argument("--lr", type=float, default=2e-4)
    ap.add_argument("--max-length", type=int, default=512)
    args = ap.parse_args()
    train(
        data_path=args.data,
        base_model_id=args.base_model,
        output_dir=args.output,
        epochs=args.epochs,
        batch_size=args.batch_size,
        lr=args.lr,
        max_length=args.max_length,
    )
