from __future__ import annotations

import re
import sys
from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import mm
from reportlab.platypus import ListFlowable, ListItem, Paragraph, SimpleDocTemplate, Spacer


def inline_markup(text: str) -> str:
    text = text.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")
    text = re.sub(r"`([^`]+)`", r"<font name='Courier'>\1</font>", text)
    return text


def build_pdf(source: Path, target: Path) -> None:
    styles = getSampleStyleSheet()
    styles.add(
        ParagraphStyle(
            name="DocTitle",
            parent=styles["Title"],
            fontName="Helvetica-Bold",
            fontSize=22,
            leading=28,
            textColor=colors.HexColor("#1f3b22"),
            spaceAfter=8,
        )
    )
    styles.add(
        ParagraphStyle(
            name="DocH2",
            parent=styles["Heading2"],
            fontName="Helvetica-Bold",
            fontSize=14,
            leading=18,
            textColor=colors.HexColor("#244b2a"),
            spaceBefore=12,
            spaceAfter=6,
        )
    )
    styles.add(
        ParagraphStyle(
            name="DocBody",
            parent=styles["BodyText"],
            fontName="Helvetica",
            fontSize=9.6,
            leading=13.6,
            spaceAfter=6,
        )
    )
    styles.add(
        ParagraphStyle(
            name="DocBullet",
            parent=styles["BodyText"],
            fontName="Helvetica",
            fontSize=9.2,
            leading=12.8,
            leftIndent=0,
            spaceAfter=3,
        )
    )
    styles.add(
        ParagraphStyle(
            name="DocNumber",
            parent=styles["BodyText"],
            fontName="Helvetica",
            fontSize=9.2,
            leading=12.8,
            leftIndent=0,
            spaceAfter=3,
        )
    )

    story = []
    bullet_items = []
    number_items = []

    def flush_lists() -> None:
        nonlocal bullet_items, number_items
        if bullet_items:
            story.append(ListFlowable(bullet_items, bulletType="bullet", leftIndent=14, bulletFontName="Helvetica", bulletFontSize=8))
            story.append(Spacer(1, 2 * mm))
            bullet_items = []
        if number_items:
            story.append(ListFlowable(number_items, bulletType="1", start="1", leftIndent=18, bulletFontName="Helvetica", bulletFontSize=8))
            story.append(Spacer(1, 2 * mm))
            number_items = []

    for raw_line in source.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line:
            flush_lists()
            continue
        if line.startswith("# "):
            flush_lists()
            story.append(Paragraph(inline_markup(line[2:]), styles["DocTitle"]))
            continue
        if line.startswith("## "):
            flush_lists()
            story.append(Paragraph(inline_markup(line[3:]), styles["DocH2"]))
            continue
        if line.startswith("- "):
            bullet_items.append(ListItem(Paragraph(inline_markup(line[2:]), styles["DocBullet"])))
            continue
        numbered = re.match(r"^\d+\.\s+(.*)$", line)
        if numbered:
            number_items.append(ListItem(Paragraph(inline_markup(numbered.group(1)), styles["DocNumber"])))
            continue
        flush_lists()
        story.append(Paragraph(inline_markup(line), styles["DocBody"]))

    flush_lists()
    target.parent.mkdir(parents=True, exist_ok=True)
    doc = SimpleDocTemplate(
        str(target),
        pagesize=A4,
        rightMargin=18 * mm,
        leftMargin=18 * mm,
        topMargin=16 * mm,
        bottomMargin=16 * mm,
        title=source.stem,
        author="Codex",
    )
    doc.build(story)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("Usage: render_markdown_pdf.py <source.md> <target.pdf>")
    build_pdf(Path(sys.argv[1]), Path(sys.argv[2]))
