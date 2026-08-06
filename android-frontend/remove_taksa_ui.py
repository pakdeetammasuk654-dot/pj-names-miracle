import re
with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# Remove Step 4: Taksa Target Focus Priority
content = re.sub(
    r'\s*// Step 4: Taksa Target Focus Priority\s*Text\(\s*text = "4\. เลือกเป้าหมายสิริมงคลที่ต้องการเน้น \(\#1\)",.*?\}\s*\}\s*\}\s*',
    '\n\n            ',
    content,
    flags=re.DOTALL
)

# Rename Step 5 to 4
content = content.replace(
    'text = "5. คำอธิษฐานมงคลพิเศษ (AI Semantic Wish)"',
    'text = "4. คำอธิษฐานมงคลพิเศษ (AI Semantic Wish)"'
)
content = content.replace(
    '// Step 5: AI Wish Query Text & Presets',
    '// Step 4: AI Wish Query Text & Presets'
)

with open("app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
