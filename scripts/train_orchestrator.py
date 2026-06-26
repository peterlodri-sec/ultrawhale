#!/usr/bin/env python3
"""Fine-tune kompress-superpower-orchestrator — DoRA + NEFTune + 4-bit.

SOTA upgrades over v1:
- DoRA (Weight-Decomposed LoRA): splits weights into magnitude+direction
- NEFTune: adds noise to embeddings during training for better chat quality
- 4-bit NF4 quantization
"""

import argparse, json, logging, torch
from pathlib import Path
from datasets import Dataset
from transformers import AutoTokenizer, AutoModelForCausalLM, TrainingArguments, Trainer, BitsAndBytesConfig
from peft import LoraConfig, get_peft_model, TaskType
from torch.utils.data import DataLoader

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)
BASE_MODEL = "Qwen/Qwen2.5-7B-Instruct"


class NEFTuneCallback:
    """Adds noise to input embeddings during training for better chat quality.
    Reference: https://arxiv.org/abs/2310.05914
    """
    def __init__(self, noise_alpha: float = 5.0):
        self.noise_alpha = noise_alpha
        self._original_embed = None
    
    def on_step_begin(self, args, state, control, **kwargs):
        model = kwargs.get("model")
        if model is None:
            return
        # Find embedding layer
        embed = model.get_input_embeddings()
        if self._original_embed is None:
            self._original_embed = embed.weight.data.clone()
        # Add noise scaled by alpha/√L
        noise = torch.randn_like(embed.weight.data) * (self.noise_alpha / (768 ** 0.5))
        embed.weight.data.add_(noise)
    
    def on_step_end(self, args, state, control, **kwargs):
        model = kwargs.get("model")
        if model is None or self._original_embed is None:
            return
        model.get_input_embeddings().weight.data = self._original_embed.clone()


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data/orchestrator_train.jsonl")
    ap.add_argument("--output", default="kompress-superpower-orchestrator")
    ap.add_argument("--epochs", type=int, default=3)
    ap.add_argument("--lr", type=float, default=2e-4)
    ap.add_argument("--max-length", type=int, default=1024)
    ap.add_argument("--neftune-alpha", type=float, default=5.0)
    args = ap.parse_args()

    log.info("Tokenizing...")
    tokenizer = AutoTokenizer.from_pretrained(BASE_MODEL, trust_remote_code=True)
    tokenizer.pad_token = tokenizer.eos_token

    with open(args.data) as f:
        rows = [json.loads(l) for l in f]
    texts = [tokenizer.apply_chat_template(r["messages"], tokenize=False, add_generation_prompt=False) for r in rows]
    encodings = tokenizer(texts, truncation=True, max_length=args.max_length, padding="max_length")
    encodings["labels"] = [ids[:] for ids in encodings["input_ids"]]
    dataset = Dataset.from_dict(encodings)
    log.info("Loaded %d pairs", len(dataset))

    log.info("Loading model (4-bit NF4)...")
    bnb = BitsAndBytesConfig(load_in_4bit=True, bnb_4bit_compute_dtype=torch.bfloat16, bnb_4bit_use_double_quant=True, bnb_4bit_quant_type="nf4")
    model = AutoModelForCausalLM.from_pretrained(BASE_MODEL, quantization_config=bnb, device_map="auto", trust_remote_code=True)

    # DoRA: Weight-Decomposed Low-Rank Adaptation (better than LoRA)
    model = get_peft_model(model, LoraConfig(
        r=16, lora_alpha=32,
        target_modules=["q_proj","k_proj","v_proj","o_proj","gate_proj","up_proj","down_proj"],
        lora_dropout=0.05, bias="none", task_type=TaskType.CAUSAL_LM,
        use_dora=True,  # ← Weight-Decomposed LoRA
    ))
    model.print_trainable_parameters()
    model.config.use_cache = False
    model.gradient_checkpointing_enable()

    neftune = NEFTuneCallback(noise_alpha=args.neftune_alpha)

    trainer = Trainer(
        model=model,
        args=TrainingArguments(
            output_dir=args.output, num_train_epochs=args.epochs,
            per_device_train_batch_size=1, gradient_accumulation_steps=16,
            learning_rate=args.lr, warmup_ratio=0.05, logging_steps=5,
            save_strategy="epoch", bf16=True, report_to="none",
            remove_unused_columns=False,
        ),
        train_dataset=dataset,
        callbacks=[neftune],
    )

    log.info("Training %d epochs (DoRA + NEFTune α=%.0f)...", args.epochs, args.neftune_alpha)
    trainer.train()

    out_dir = Path(args.output)
    out_dir.mkdir(exist_ok=True)
    model.save_pretrained(str(out_dir))
    tokenizer.save_pretrained(str(out_dir))
    log.info("Saved to %s", out_dir)


if __name__ == "__main__":
    main()
