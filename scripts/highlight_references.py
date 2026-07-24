from pathlib import Path

import fitz


ROOT = Path(__file__).resolve().parents[1]
REFERENCE_DIR = ROOT / "referensi"
OUTPUT_DIR = REFERENCE_DIR / "highlighted"

YELLOW = (1.0, 0.92, 0.0)
RED = (1.0, 0.25, 0.25)


REFERENCES = [
    {
        "file": "admin,+F-07.pdf",
        "yellow_points": [
            "bahasa alami",
            "sistem pengolah bahasa alami",
            "aturan produksi",
            "kalimat tanya",
        ],
        "red_points": [
            "scanner",
            "parser",
            "translator",
            "evaluator",
            "pohon sintaks",
        ],
    },
    {
        "file": "caldarini-2022-literature-survey-chatbots.pdf",
        "yellow_points": [
            "chatbot is a software application",
            "Natural Language Processing",
            "Artificial Intelligence",
            "human-like conversation",
        ],
        "red_points": [
            "pattern matching",
            "rule-based",
            "retrieval-based",
            "generative",
            "rule based",
        ],
    },
    {
        "file": "chandrakala-2023-intent-recognition-pipeline.pdf",
        "yellow_points": [
            "conversational AI",
            "Natural Language Understanding",
            "pipeline",
        ],
        "red_points": [
            "intent recognition",
            "intent classification",
            "entity extraction",
        ],
    },
    {
        "file": "warto-2024-systematic-literature-review-ner.pdf",
        "yellow_points": [
            "Systematic Literature Review",
            "Natural Language Processing",
            "question answering",
        ],
        "red_points": [
            "Named Entity Recognition",
            "entity extraction",
            "named entities",
        ],
    },
    {
        "file": "martinengo-2023-conversational-agents-health-care.pdf",
        "yellow_points": [
            "health care",
            "conversational agents",
            "classification",
            "conceptual framework",
        ],
        "red_points": [
            "rule-based",
            "safety",
            "privacy",
            "security",
            "ethical",
        ],
    },
    {
        "file": "klug-2024-clinical-nlp-patient-journey.pdf",
        "yellow_points": [
            "patient journey",
            "clinical natural language processing",
            "clinical NLP",
            "systematic review",
        ],
        "red_points": [
            "information extraction",
            "admission",
            "discharge",
        ],
    },
]


def normalize_text(text):
    return " ".join(text.lower().split())


def line_rect(line):
    rect = fitz.Rect(line["bbox"])
    return rect + (-1, -1, 1, 1)


def highlight_line(page, line, color):
    annot = page.add_highlight_annot(line_rect(line))
    annot.set_colors(stroke=color)
    annot.set_opacity(0.45 if color == YELLOW else 0.55)
    annot.update()


def collect_sentence_lines(lines, index):
    selected = [index]

    cursor = index
    while cursor > 0:
        previous_text = lines[cursor - 1]["text"].strip()
        current_text = lines[cursor]["text"].strip()
        if not previous_text or previous_text.endswith((".", "?", "!", ":")):
            break
        if current_text and current_text[0].isupper():
            break
        selected.insert(0, cursor - 1)
        cursor -= 1

    cursor = index
    while cursor + 1 < len(lines):
        current_text = lines[cursor]["text"].strip()
        next_text = lines[cursor + 1]["text"].strip()
        if not next_text:
            break
        if current_text.endswith((".", "?", "!", ":")):
            break
        selected.append(cursor + 1)
        cursor += 1

    return selected


def extract_page_lines(page):
    page_lines = []
    text = page.get_text("dict", flags=fitz.TEXTFLAGS_TEXT)

    for block in text.get("blocks", []):
        for line in block.get("lines", []):
            line_text = "".join(span.get("text", "") for span in line.get("spans", []))
            if line_text.strip():
                page_lines.append({"text": line_text, "bbox": line["bbox"]})

    return page_lines


def add_point_highlights(doc, phrases, color, max_hits_per_phrase):
    total = 0
    highlighted = set()

    for phrase in phrases:
        phrase_count = 0
        needle = normalize_text(phrase)

        for page_index, page in enumerate(doc):
            if phrase_count >= max_hits_per_phrase:
                break

            lines = extract_page_lines(page)
            for line_index, line in enumerate(lines):
                if phrase_count >= max_hits_per_phrase:
                    break
                if needle not in normalize_text(line["text"]):
                    continue

                for selected_index in collect_sentence_lines(lines, line_index):
                    key = (page_index, selected_index, color)
                    if key in highlighted:
                        continue
                    highlight_line(page, lines[selected_index], color)
                    highlighted.add(key)
                    total += 1

                phrase_count += 1

    return total


def highlight_pdf(reference):
    source = REFERENCE_DIR / reference["file"]
    output = OUTPUT_DIR / reference["file"].replace(".pdf", "-highlighted.pdf")

    if not source.exists():
        return reference["file"], 0, 0, "missing"

    doc = fitz.open(source)
    yellow_count = 0
    red_count = 0

    yellow_count += add_point_highlights(doc, reference["yellow_points"], YELLOW, 3)
    red_count += add_point_highlights(doc, reference["red_points"], RED, 4)

    doc.save(output, garbage=4, deflate=True)
    doc.close()

    return output.name, yellow_count, red_count, "ok"


def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    print("Highlight PDF referensi")
    print("Kuning: kalimat penting umum")
    print("Merah : kalimat yang dipakai langsung di laporan")
    print()

    for reference in REFERENCES:
        filename, yellow_count, red_count, status = highlight_pdf(reference)
        print(f"{status:7} {filename}")
        print(f"        yellow={yellow_count} red={red_count}")


if __name__ == "__main__":
    main()
