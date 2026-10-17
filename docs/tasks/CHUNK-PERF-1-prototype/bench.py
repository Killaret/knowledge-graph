"""CHUNK-PERF-1: timings of chunk() on long paragraphs.

    python docs/tasks/CHUNK-PERF-1-prototype/bench.py --module nlp-service/app/core/chunking.py
    python docs/tasks/CHUNK-PERF-1-prototype/bench.py --module ... --tokenizer <path to MiniLM snapshot>

Without --tokenizer the counter is the word count used by test_chunking.py. With it, the real
MiniLM tokenizer (add_special_tokens=False, as _chunk_token_counter) — slower, and the only way
to see the _char_spans cost on unbreakable text.
"""
import argparse
import base64
import importlib.util
import random
import re
import sys
import time


def load(path):
    spec = importlib.util.spec_from_file_location("chunking_bench", path)
    module = importlib.util.module_from_spec(spec)
    sys.modules["chunking_bench"] = module
    spec.loader.exec_module(module)
    return module


def prose(size):
    rnd = random.Random(1)
    words = "альфа бета гамма дельта эпсилон текст пример см. т.е. 3.14 А.С. www.site.ru".split()
    out = []
    while sum(map(len, out)) < size:
        out.append(" ".join(rnd.choice(words) for _ in range(rnd.randint(5, 15))).capitalize() + ".")
    return " ".join(out)[:size]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--module", required=True)
    ap.add_argument("--tokenizer")
    args = ap.parse_args()
    chunking = load(args.module)
    rnd = random.Random(2)
    if args.tokenizer:
        from transformers import AutoTokenizer

        tok = AutoTokenizer.from_pretrained(args.tokenizer, local_files_only=True)
        count = lambda s: len(tok.encode(s, add_special_tokens=False))
        cjk = "".join(rnd.choice("的一是不了人我在有他这中大来上国个到说们为子和你地出道也时年得就那要下以生会自着去") for _ in range(60_000))
        cases = [
            ("cjk 60K", cjk),
            ("base64 60K", base64.b64encode(bytes(rnd.getrandbits(8) for _ in range(45_000))).decode()[:60_000]),
            ("1. 60K", "1." * 30_000),
            ("prose 100K", prose(100_000)),
        ]
    else:
        count = lambda s: len(re.findall(r"\w+", s))
        cases = [
            ("prose 400K", prose(400_000)),
            ("www. 400K", "www." * 100_000),
            ("1. 400K", "1." * 200_000),
            ("a. 400K", "a." * 200_000),
            ("1. 800K", "1." * 400_000),
        ]
    params = chunking.ChunkingParams(token_counter=count, target_tokens=126, max_tokens=126)
    for name, text in cases:
        started = time.perf_counter()
        chunks = chunking.chunk(text, params)
        print(f"{name:12} {time.perf_counter() - started:7.2f}s  chunks={len(chunks)}", flush=True)


if __name__ == "__main__":
    main()
