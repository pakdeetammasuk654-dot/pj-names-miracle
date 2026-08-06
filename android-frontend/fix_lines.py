with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    lines = f.readlines()

new_lines = [
    '                    // Universal Blessing Text\n',
    '                    Text(\n',
    '                        text = "📝 คำอธิษฐานมงคลประจำวัน",\n',
    '                        color = Color.LightGray,\n',
    '                        fontSize = 12.sp\n',
    '                    )\n',
    '                    Text(\n',
    '                        text = "\\"\\$wishQuery\\"",\n',
    '                        color = GoldLight,\n',
    '                        fontSize = 13.sp,\n',
    '                        textAlign = TextAlign.Center\n',
    '                    )\n',
    '\n',
    '                    Spacer(modifier = Modifier.height(12.dp))\n',
    '\n',
    '                    Text(\n',
    '                        text = "✨ คำอวยพรมงคลชีวิตประจำนาม",\n',
    '                        color = Color.LightGray,\n',
    '                        fontSize = 12.sp\n',
    '                    )\n',
    '                    Text(\n',
    '                        text = "\\"ขอให้นามมงคล \'${item.thName}\' นำพาชีวิตให้เปี่ยมด้วยพลังแห่งความสุข ปัญญา และความเจริญรุ่งเรือง ก้าวหน้าประสบความสำเร็จในชีวิตสืบไป\\"",\n',
    '                        color = Color.White,\n',
    '                        fontSize = 13.sp,\n',
    '                        textAlign = TextAlign.Center,\n',
    '                        fontWeight = FontWeight.Medium\n',
    '                    )\n',
    '                }\n',
    '            }\n'
]

# lines[218:247] replaces lines 219 to 247 inclusive.
lines[218:247] = new_lines

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.writelines(lines)
