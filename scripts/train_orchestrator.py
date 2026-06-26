#!/usr/bin/env python3
"""Fine-tune kompress-superpower-orchestrator on experiment history."""
import argparse, json, logging, torch
from pathlib import Path
from datasets import Dataset
from transformers import AutoTokenizer, AutoModelForCausalLM, TrainingArguments, Trainer
from transformers import DataCollatorForSeq2Seq  # handles padding properly
from peft import LoraConfig, get_peft_model, TaskType

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)
BASE_MODEL = "Qwen/Qwen2.5-7B-Instruct"

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data/orchestrator_train.jsonl")
    ap.add_argument("--output", default="kompress-superpower-orchestrator")
    ap.add_argument("--epochs", type=int, default=3)
    ap.add_argument("--batch-size", type=int, default=1)
    ap.add_argument("--lr", type=float, default=2e-4)
    ap.add_argument("--max-length", type=int, default=1024)
    args = ap.parse_args()

    log.info("Tokenizing %s...", BASE_MODEL)
    tokenizer = AutoTokenizer.from_pretrained(BASE_MODEL, trust_remote_code=True)
    tokenizer.pad_token = tokenizer.eos_token

    with open(args.data) as f:
        rows = [json.loads(l) for l in f]
    
    texts = []
    for r in rows:
        t = tokenizer.apply_chat_template(r["messages"], tokenize=False, add_generation_prompt=False)
        texts.append(t)
    
    # Tokenize with padding to max_length
    encodings = tokenizer(texts, truncation=True, max_length=args.max_length, padding="max_length")
    encodings["labels"] = [ids[:] for ids in encodings["input_ids"]]  # copy for labels
    dataset = Dataset.from_dict(encodings)
    log.info("Loaded %d pairs", len(dataset))

    log.info("Loading model...")
    from transformers import BitsAndBytesConfig
    bnb_config = BitsAndBytesConfig(load_in_4bit=True, bnb_4bit_compute_dtype=torch.bfloat16, bnb_4bit_use_double_quant=True, bnb_4bit_quant_type="nf4")
    model = AutoModelForCausalLM.from_pretrained(BASE_MODEL, quantization_config=bnb_config, device_map="auto", trust_remote_code=True)
    model = get_peft_model(model, LoraConfig(
        r=16, lora_alpha=32,
        target_modules=["q_proj","k_proj","v_proj","o_proj","gate_proj","up_proj","down_proj"],
        lora_dropout=0.05, bias="none", task_type=TaskType.CAUSAL_LM,
    ))
    model.print_trainable_parameters()
    model.config.use_cache = False
    model.gradient_checkpointing_enable()
    model.enable_input_require_grads()  # for gradient checkpointing

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
