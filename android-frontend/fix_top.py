with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "r", encoding="utf-8") as f:
    lines = f.readlines()

for i, line in enumerate(lines):
    if "val wishQuery =" in line and lines[i+1].strip().startswith('"犧'):
        lines[i+1] = '        "ขอให้นามมงคลนี้นำพาความสุข ความเจริญรุ่งเรือง สติปัญญา และความสำเร็จมาสู่ชีวิต"\n'
    elif "contentDescription =" in line and "ArrowBack" in lines[i-1]:
        lines[i] = '                        contentDescription = "ย้อนกลับ",\n'
    elif "text =" in line and "笨" in line:
        lines[i] = '                            text = "✨ การ์ดคำอวยพรสิริมงคลประจำนาม ✨",\n'
    elif "text =" in line and "側" in line:
        lines[i] = '                        text = "👤 ชื่อมงคลที่ได้รับ",\n'
    elif "text =" in line and "当" in line:
        lines[i] = '                        text = "💡 ความหมายอันเป็นมงคล",\n'
    elif "text = item.meaning.ifEmpty" in line:
        lines[i] = '                        text = item.meaning.ifEmpty { "ผู้มีความสุขความเจริญ สติปัญญา และมั่งคั่งสมบูรณ์พร้อม" },\n'
    elif "text =" in line and "套" in line:
        lines[i] = '                            text = "📅 วันเกิด: ${formatDayThai(dayOfBirth)}  [✔️ ปลอดกาลกิณี 100%]",\n'

with open("app/src/main/java/com/miracle/names/ui/blessing/BlessingCardScreen.kt", "w", encoding="utf-8") as f:
    f.writelines(lines)
