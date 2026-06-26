#!/usr/bin/env python3
"""Fine-tune kompress-superpower-orchestrator on experiment history.

LoRA fine-tune of Qwen2.5-7B-Instruct on ~100-200 function-calling
conversation pairs encoding 17 kompress experiment outcomes.

Usage (on vast.ai GPU):
  pip install -q transformers peft datasets torch accelerate bitsandbytes
  python scripts/train_orchestrator.py \
    --data data/orchestrator_train.jsonl \
    --output kompress-superpower-orchestrator \
    --epochs 3
"""

import argparse, json, logging, torch
from pathlib import Path
from datasets import Dataset
from transformers import AutoTokenizer, AutoModelForCausalLM, TrainingArguments, Trainer
from peft import LoraConfig, get_peft_model, TaskType

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

BASE_MODEL = "Qwen/Qwen2.5-7B-Instruct"


def load_data(path: str, tokenizer, max_length: int = 1024):
    """Load conversation pairs and tokenize."""
    with open(path) as f:
        rows = [json.loads(l) for l in f]
    
    texts = []
    for row in rows:
        msgs = row["messages"]
        text = tokenizer.apply_chat_template(
            msgs,
            tokenize=False,
            add_generation_prompt=False,
        )
        texts.append(text)
    
    dataset = Dataset.from_dict({"text": texts})
    
    def tokenize(examples):
        return tokenizer(
            examples["text"],
            truncation=True,
            max_length=max_length,
            padding=False,
        )
    
    dataset = dataset.map(tokenize, remove_columns=["text"])
    log.info("Loaded %d pairs", len(dataset))
    return dataset


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data/orchestrator_train.jsonl")
    ap.add_argument("--output", default="kompress-superpower-orchestrator")
    ap.add_argument("--epochs", type=int, default=3)
    ap.add_argument("--batch-size", type=int, default=2)
    ap.add_argument("--lr", type=float, default=2e-4)
    ap.add_argument("--max-length", type=int, default=1024)
    ap.add_argument("--lora-r", type=int, default=16)
    ap.add_argument("--lora-alpha", type=int, default=32)
    args = ap.parse_args()

    log.info("Loading %s...", BASE_MODEL)
    tokenizer = AutoTokenizer.from_pretrained(BASE_MODEL, trust_remote_code=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    model = AutoModelForCausalLM.from_pretrained(
        BASE_MODEL,
        torch_dtype=torch.bfloat16,
        device_map="auto",
        trust_remote_code=True,
    )

    # LoRA
    lora_config = LoraConfig(
        r=args.lora_r,
        lora_alpha=args.lora_alpha,
        target_modules=["q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"],
        lora_dropout=0.05,
        bias="none",
        task_type=TaskType.CAUSAL_LM,
    )
    model = get_peft_model(model, lora_config)
    model.print_trainable_parameters()

    # Data
    dataset = load_data(args.data, tokenizer, args.max_length)

    # Train
    training_args = TrainingArguments(
        output_dir=args.output,
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch_size,
        gradient_accumulation_steps=4,
        learning_rate=args.lr,
        warmup_ratio=0.05,
        logging_steps=5,
        save_strategy="epoch",
        bf16=True,
        report_to="none",
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=dataset,
        tokenizer=tokenizer,
    )

    log.info("Training %d epochs...", args.epochs)
    trainer.train()

    # Save
    out_dir = Path(args.output)
    out_dir.mkdir(exist_ok=True)
    model.save_pretrained(str(out_dir))
    tokenizer.save_pretrained(str(out_dir))
    log.info("Saved to %s", out_dir)


if __name__ == "__main__":
    main()
