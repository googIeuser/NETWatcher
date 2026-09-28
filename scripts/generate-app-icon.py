"""Generate the NetWatcher PNG and multi-size Windows icon.

Requires Pillow: python -m pip install Pillow
"""

from pathlib import Path

from PIL import Image, ImageDraw


ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "flutter_app" / "assets"
INK = (23, 40, 33, 255)
SIGNAL = (217, 244, 106, 255)
SIZE = 1024


def draw_icon() -> Image.Image:
    image = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)

    # The bright control-desk tile keeps the mark recognizable at 16 px.
    draw.rounded_rectangle((40, 40, 984, 984), radius=184, fill=SIGNAL)

    # A compact, geometric N with overlapping joins instead of font glyphs.
    draw.rectangle((244, 228, 370, 796), fill=INK)
    draw.polygon(((344, 228), (474, 228), (680, 796), (550, 796)), fill=INK)
    draw.rectangle((654, 228, 780, 796), fill=INK)
    return image


def main() -> None:
    image = draw_icon()
    ASSETS.mkdir(parents=True, exist_ok=True)
    image.resize((512, 512), Image.Resampling.LANCZOS).save(ASSETS / "app_icon.png")
    image.save(
        ASSETS / "app_icon.ico",
        format="ICO",
        sizes=[(size, size) for size in (16, 24, 32, 48, 64, 128, 256)],
    )


if __name__ == "__main__":
    main()
