from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter, ImageFont, ImageOps


ROOT = Path(__file__).resolve().parents[3]
OUTPUT_DIR = Path(__file__).resolve().parent
BACKGROUND = Image.open(OUTPUT_DIR / "emerald-gold-background-v1.png").convert("RGB")
FONT_TTC = "/System/Library/Fonts/Supplemental/SukhumvitSet.ttc"
WIDTH, HEIGHT = 1242, 2688
PHONE_WIDTH = 982
PHONE_X = (WIDTH - PHONE_WIDTH) // 2
PHONE_Y = 555

SCREENSHOTS = [
    ("iphone-65-01.png", "ชื่อดี 4 มิติ", "วิเคราะห์ชื่อมงคลเจาะลึก 4 ด้าน", "เลขศาสตร์ • พลังเงา • ความหมาย • ความไพเราะ"),
    ("iphone-65-02.png", "ปรับสูตรมงคล 4 มิติ", "จัดอันดับชื่อที่เหมาะกับคุณที่สุด", "ปรับน้ำหนักทั้ง 4 มิติได้อิสระ 0% - 100%"),
    ("iphone-65-03.png", "การันตีมงคล 100%", "ไร้อักษรกาลกิณี พร้อมการ์ดอวยพร", "สร้างการ์ดมงคลและวอลเปเปอร์เสริมดวงเฉพาะคุณ"),
]

CURRENT_SCREENS = [
    (
        Path("/Users/tayap/Desktop/Screenshot 2569-08-14 at 16.28.28.png"),
        "ชื่อดี 4 มิติ",
        "วิเคราะห์ชื่อมงคลเจาะลึก 4 ด้าน",
        "เลขศาสตร์ • พลังเงา • ความหมาย • ความไพเราะ",
    ),
    (
        Path("/Users/tayap/Desktop/Screenshot 2569-08-14 at 16.28.53.png"),
        "ปรับสูตรมงคล 4 มิติ",
        "จัดอันดับชื่อที่เหมาะกับคุณที่สุด",
        "ปรับน้ำหนักทั้ง 4 มิติได้อิสระ 0% - 100%",
    ),
    (
        Path("/Users/tayap/Desktop/Screenshot 2569-08-14 at 16.29.06.png"),
        "การันตีมงคล 100%",
        "ไร้อักษรกาลกิณี พร้อมการ์ดอวยพร",
        "สร้างการ์ดมงคลและวอลเปเปอร์เสริมดวงเฉพาะคุณ",
    ),
]


def rounded(image: Image.Image, radius: int) -> Image.Image:
    mask = Image.new("L", image.size, 0)
    ImageDraw.Draw(mask).rounded_rectangle((0, 0, image.width, image.height), radius, fill=255)
    result = Image.new("RGBA", image.size)
    result.paste(image, (0, 0), mask)
    return result


def crop_black_canvas(image: Image.Image) -> Image.Image:
    """Remove the empty black desktop canvas while retaining the phone itself."""
    rgb = image.convert("RGB")
    visible = rgb.convert("L").point(lambda pixel: 255 if pixel > 12 else 0)
    bounds = visible.getbbox()
    return rgb.crop(bounds) if bounds else rgb


def create_screenshot(index: int, source_path: Path, title: str, subtitle: str, detail: str, output_name: str) -> None:
    canvas = ImageOps.fit(BACKGROUND, (WIDTH, HEIGHT), method=Image.Resampling.LANCZOS, centering=(0.5, 0.22)).convert("RGBA")
    canvas = Image.alpha_composite(canvas, Image.new("RGBA", (WIDTH, HEIGHT), (0, 21, 12, 82)))

    panel = Image.new("RGBA", (WIDTH - 80, 405), (4, 34, 21, 190))
    panel_mask = Image.new("L", panel.size, 0)
    ImageDraw.Draw(panel_mask).rounded_rectangle((0, 0, panel.width, panel.height), 42, fill=255)
    panel.putalpha(panel_mask)
    canvas.alpha_composite(panel, (40, 62))

    draw = ImageDraw.Draw(canvas)
    gold, white, mint = (243, 229, 171, 255), (255, 252, 240, 255), (184, 236, 198, 255)
    draw.rounded_rectangle((82, 110, 285, 162), 26, outline=(222, 186, 57, 255), width=2)
    draw.text((112, 118), "ชื่อดี", font=ImageFont.truetype(FONT_BOLD, 29), fill=gold)
    draw.text((82, 194), title, font=ImageFont.truetype(FONT_BOLD, 63), fill=white)
    draw.text((84, 280), subtitle, font=ImageFont.truetype(FONT_BOLD, 39), fill=gold)
    draw.text((85, 335), detail, font=ImageFont.truetype(FONT_REGULAR, 30), fill=mint)

    source = crop_black_canvas(Image.open(source_path))
    phone_height = round(source.height * PHONE_WIDTH / source.width)
    phone = rounded(source.resize((PHONE_WIDTH, phone_height), Image.Resampling.LANCZOS), 58)

    shadow = Image.new("RGBA", (PHONE_WIDTH + 46, phone_height + 56), (0, 0, 0, 0))
    ImageDraw.Draw(shadow).rounded_rectangle((23, 20, PHONE_WIDTH + 23, phone_height + 20), 68, fill=(0, 0, 0, 145))
    canvas.alpha_composite(shadow.filter(ImageFilter.GaussianBlur(18)), (PHONE_X - 23, PHONE_Y - 20))

    border = Image.new("RGBA", (PHONE_WIDTH + 10, phone_height + 10), (0, 0, 0, 0))
    ImageDraw.Draw(border).rounded_rectangle((0, 0, PHONE_WIDTH + 9, phone_height + 9), 63, outline=(243, 229, 171, 230), width=3)
    canvas.alpha_composite(border, (PHONE_X - 5, PHONE_Y - 5))
    canvas.alpha_composite(phone, (PHONE_X, PHONE_Y))
    canvas.convert("RGB").save(OUTPUT_DIR / output_name, quality=95)


for number, screenshot in enumerate(SCREENSHOTS, 1):
    create_screenshot(
        number,
        ROOT / "ios-frontend/AppStoreScreenshots" / screenshot[0],
        *screenshot[1:],
        f"iphone-65-marketing-{number:02d}.png",
    )

for number, screenshot in enumerate(CURRENT_SCREENS, 1):
    create_screenshot(number, *screenshot, f"iphone-65-current-{number:02d}.png")
