"""Generate the NetWatcher PNG and multi-size Windows icon.

Requires Pillow: python -m pip install Pillow
"""

from pathlib import Path

from PIL import Image, ImageDraw


ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "flutter_app" / "assets"
SIZE = 1024


def draw_icon() -> Image.Image:
    image = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)

    # The four endpoints and central node mirror the network-map dashboard.
    draw.rounded_rectangle((34, 34, 990, 990), radius=218, fill="#F0F0E5")
    links = [
        ((280, 342), (512, 506)),
        ((744, 310), (512, 506)),
        ((270, 710), (512, 506)),
        ((750, 718), (512, 506)),
    ]
    for start, end in links:
        draw.line((start, end), fill="#297A69", width=72)
    for x, y in ((280, 342), (744, 310), (270, 710), (750, 718)):
        draw.ellipse((x - 85, y - 85, x + 85, y + 85), fill="#172821")
        draw.ellipse((x - 27, y - 27, x + 27, y + 27), fill="#D9F46A")
    draw.ellipse(
        (383, 377, 641, 635),
        fill="#D9F46A",
        outline="#172821",
        width=48,
    )
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
