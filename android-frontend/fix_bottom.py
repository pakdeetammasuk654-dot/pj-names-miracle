import re
with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

content = re.sub(
    r'text = "町.*?",\n\s*color = Color\.LightGray,\n\s*fontSize = 12\.sp',
    r'''text = "📝 คำอธิษฐานมงคลประจำวัน",
                        color = Color.LightGray,
                        fontSize = 12.sp''',
    content,
    flags=re.DOTALL
)

content = re.sub(
    r'text = "糖.*?",\n\s*color = Color\.LightGray,\n\s*fontSize = 12\.sp',
    r'''text = "✨ คำอวยพรมงคลชีวิตประจำนาม",
                        color = Color.LightGray,
                        fontSize = 12.sp''',
    content,
    flags=re.DOTALL
)

content = re.sub(
    r'text = "\"犧.*?"",\n\s*color = Color\.White,',
    r'''text = "\"ขอให้นามมงคล '${item.thName}' นำพาชีวิตให้เปี่ยมด้วยพลังแห่งความสุข ปัญญา และความเจริญรุ่งเรือง ก้าวหน้าประสบความสำเร็จในชีวิตสืบไป\"",
                        color = Color.White,''',
    content,
    flags=re.DOTALL
)

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
