#!/usr/bin/env python3
"""Fine-tune kompress-superpower-orchestrator on experiment history."""
import argparse, json, logging, torch
from pathlib import Path
from datasets import Dataset
from transformers import AutoTokenizer, AutoModelForCausalLM, TrainingArguments, Trainer, DataCollatorForLanguageModeling
from peft import LoraConfig, get_peft_model, TaskType

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)
BASE_MODEL = "Qwen/Qwen2.5-7B-Instruct"

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data/orchestrator_train.jsonl")
    ap.add_argument("--output", default="kompress-superpower-orchestrator")
    ap.add_argument("--epochs", type=int, default=3)
    ap.add_argument("--batch-size", type=int, default=2)
    ap.add_argument("--lr", type=float, default=2e-4)
    ap.add_argument("--max-length", type=int, default=1024)
    args = ap.parse_args()

    log.info("Loading %s...", BASE_MODEL)
    tokenizer = AutoTokenizer.from_pretrained(BASE_MODEL, trust_remote_code=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    model = AutoModelForCausalLM.from_pretrained(BASE_MODEL, torch_dtype=torch.bfloat16, device_map="auto", trust_remote_code=True)
    model = get_peft_model(model, LoraConfig(r=16, lora_alpha=32, target_modules=["q_proj","k_proj","v_proj","o_proj","gate_proj","up_proj","down_proj"], lora_dropout=0.05, bias="none", task_type=TaskType.CAUSAL_LM))
    model.print_trainable_parameters()

    with open(args.data) as f:
        rows = [json.loads(l) for l in f]
    texts = [tokenizer.apply_chat_template(r["messages"], tokenize=False, add_generation_prompt=False) for r in rows]
    
    def tokenize(ex):
        tok = tokenizer(ex["text"], truncation=True, max_length=args.max_length, padding=False)
        tok["labels"] = tok["input_ids"].copy()
        return tok
    
    dataset = Dataset.from_dict({"text": texts}).map(tokenize, remove_columns=["text"])
    log.info("Loaded %d pairs", len(dataset))

    trainer = Trainer(
        model=model,
        args=TrainingArguments(output_dir=args.output, num_train_epochs=args.epochs, per_device_train_batch_size=args.batch_size, gradient_accumulation_steps=4, learning_rate=args.lr, warmup_ratio=0.05, logging_steps=5, save_strategy="epoch", bf16=True, report_to="none"),
        train_dataset=dataset,
        data_collator=DataCollatorForLanguageModeling(tokenizer=tokenizer, mlm=False),
    )

    log.info("Training %d epochs...", args.epochs)
    trainer.train()

    out_dir = Path(args.output)
    out_dir.mkdir(exist_ok=True)
    model.save_pretrained(str(out_dir))
    tokenizer.save_pretrained(str(out_dir))
    log.info("Saved to %s", out_dir)

if __name__ == "__main__":
    main()
