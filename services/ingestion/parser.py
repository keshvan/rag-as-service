import os

import fitz
from docx import Document


def extract_text(file_path: str) -> str:
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"File not found: {file_path}")

    lower_path = file_path.lower()
    text = ""

    if lower_path.endswith(".pdf"):
        document = fitz.open(file_path)
        try:
            for page in document:
                text += page.get_text()
        finally:
            document.close()
    elif lower_path.endswith(".txt"):
        with open(file_path, "r", encoding="utf-8") as file:
            text = file.read()
    elif lower_path.endswith(".docx"):
        document = Document(file_path)
        for paragraph in document.paragraphs:
            text += paragraph.text + "\n"
    else:
        raise ValueError(f"Unsupported file type: {file_path}")

    if not text.strip():
        raise ValueError("File is empty or could not be parsed")

    return text.strip()