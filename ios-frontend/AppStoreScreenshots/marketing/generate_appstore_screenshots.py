from pathlib import Path
from PIL import Image, ImageDraw, ImageFont, ImageOps

MARKETING_DIR = Path("ios-frontend/AppStoreScreenshots/marketing")
SCREENSHOTS_DIR = Path("ios-frontend/AppStoreScreenshots")
BG_PATH = MARKETING_DIR / "emerald-gold-background-v1.png"
background = Image.open(BG_PATH).convert("RGB")

FONT_BOLD = "ios-frontend/ChueDee/Resources/Prompt-Bold.ttf"
FONT_SEMIBOLD = "ios-frontend/ChueDee/Resources/Prompt-SemiBold.ttf"
FONT_MEDIUM = "ios-frontend/ChueDee/Resources/Prompt-Medium.ttf"

WIDTH, HEIGHT = 1242, 2688

ITEMS = [
    {
        "id": "01",
        "image": "/Users/tayap/.gemini/antigravity/brain/978581f2-6c5f-49ea-97b9-d261973c9309/.user_uploaded/media_1790398156680.png",
        "badge": "ชื่อดี 4 มิติ",
        "title": "ชื่อดี 4 มิติ",
        "subtitle": "วิเคราะห์ชื่อมงคลเจาะลึก 4 ด้าน",
        "detail": "เลขศาสตร์ • พลังเงา • ความหมาย • ความไพเราะ",
    },
    {
        "id": "02",
        "image": "/Users/tayap/.gemini/antigravity/brain/978581f2-6c5f-49ea-97b9-d261973c9309/.user_uploaded/media_1790398156666.png",
        "badge": "ชื่อดี 4 มิติ",
        "title": "ปรับสูตรมงคล 4 มิติ",
        "subtitle": "จัดอันดับชื่อที่เหมาะกับคุณที่สุด",
        "detail": "ปรับน้ำหนักทั้ง 4 มิติได้อิสระ 0% - 100%",
    },
    {
        "id": "03",
        "image": "/Users/tayap/.gemini/antigravity/brain/978581f2-6c5f-49ea-97b9-d261973c9309/.user_uploaded/media_1790398156676.png",
        "badge": "ชื่อดี 4 มิติ",
        "title": "การันตีมงคล 100%",
        "subtitle": "ไร้อักษรกาลกิณี พร้อมการ์ดอวยพร",
        "detail": "สร้างการ์ดมงคลและวอลเปเปอร์เสริมดวงเฉพาะคุณ",
    },
]

for item in ITEMS:
    # 1. Base canvas with background
    canvas = ImageOps.fit(background, (WIDTH, HEIGHT), method=Image.Resampling.LANCZOS, centering=(0.5, 0.22)).convert("RGBA")
    overlay = Image.new("RGBA", (WIDTH, HEIGHT), (0, 21, 12, 85))
    canvas = Image.alpha_composite(canvas, overlay)

    # 2. Header card
    card_w = WIDTH - 80
    card_h = 415
    panel = Image.new("RGBA", (card_w, card_h), (4, 34, 21, 195))
    panel_mask = Image.new("L", panel.size, 0)
    ImageDraw.Draw(panel_mask).rounded_rectangle((0, 0, card_w, card_h), 42, fill=255)
    panel.putalpha(panel_mask)

    # Gold stroke on card
    panel_border = Image.new("RGBA", panel.size, (0, 0, 0, 0))
    ImageDraw.Draw(panel_border).rounded_rectangle((0, 0, card_w - 1, card_h - 1), 42, outline=(222, 186, 57, 100), width=2)
    panel = Image.alpha_composite(panel, panel_border)
    canvas.alpha_composite(panel, (40, 62))

    draw = ImageDraw.Draw(canvas)
    gold = (243, 229, 171, 255)
    white = (255, 252, 240, 255)
    mint = (184, 236, 198, 255)

    # Badge auto calculation
    badge_font = ImageFont.truetype(FONT_SEMIBOLD, 28)
    bbox_badge = draw.textbbox((0, 0), item["badge"], font=badge_font)
    text_w = bbox_badge[2] - bbox_badge[0]
    text_h = bbox_badge[3] - bbox_badge[1]

    bx0 = 82
    by0 = 98
    bx1 = bx0 + text_w + 48
    by1 = by0 + 56
    draw.rounded_rectangle((bx0, by0, bx1, by1), 28, outline=(222, 186, 57, 230), width=2)
    # Center text inside badge
    tx = bx0 + (bx1 - bx0 - text_w) // 2
    ty = by0 + (by1 - by0 - text_h) // 2 - 2
    draw.text((tx, ty), item["badge"], font=badge_font, fill=gold)

    # Titles
    draw.text((82, 180), item["title"], font=ImageFont.truetype(FONT_BOLD, 64), fill=white)
    draw.text((84, 270), item["subtitle"], font=ImageFont.truetype(FONT_BOLD, 39), fill=gold)
    draw.text((85, 335), item["detail"], font=ImageFont.truetype(FONT_MEDIUM, 30), fill=mint)

    # 3. Crop and composite phone
    raw_phone = Image.open(item["image"])
    crop_bbox = (13, 46, 520, 1011)
    cropped_phone = raw_phone.crop(crop_bbox)

    TARGET_W = 1140
    TARGET_H = round(cropped_phone.height * TARGET_W / cropped_phone.width)
    phone_resized = cropped_phone.resize((TARGET_W, TARGET_H), Image.Resampling.LANCZOS)

    PHONE_X = (WIDTH - TARGET_W) // 2
    PHONE_Y = 515

    canvas.alpha_composite(phone_resized, (PHONE_X, PHONE_Y))
    result = canvas.convert("RGB")

    # Save to both marketing and root screenshots folder
    out1 = MARKETING_DIR / f"appstore-marketing-{item['id']}.png"
    out2 = MARKETING_DIR / f"iphone-65-{item['id']}.png"
    out3 = SCREENSHOTS_DIR / f"iphone-65-{item['id']}.png"

    result.save(out1, quality=95)
    result.save(out2, quality=95)
    result.save(out3, quality=95)
    print(f"Generated screenshot {item['id']}: {out1}")

