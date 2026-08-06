import re
with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# Add imports if missing
if "import androidx.compose.foundation.horizontalScroll" not in content:
    content = content.replace("import androidx.compose.foundation.lazy.itemsIndexed", "import androidx.compose.foundation.lazy.itemsIndexed\nimport androidx.compose.foundation.horizontalScroll\nimport androidx.compose.foundation.rememberScrollState")

content = re.sub(
    r'(// Badges Row\n\s*Row\()(\n\s*horizontalArrangement = Arrangement\.spacedBy\(6\.dp\),\n\s*verticalAlignment = Alignment\.CenterVertically\n\s*\))',
    r'\1modifier = Modifier.horizontalScroll(rememberScrollState()),\2',
    content
)

content = re.sub(
    r'(text = taksaDetail,\n\s*fontSize = 11\.sp,\n\s*fontWeight = FontWeight\.Bold,\n\s*color = TextGold)(\n)',
    r'\1,\n                            maxLines = 1,\n                            softWrap = false\2',
    content
)

content = re.sub(
    r'(text = "เลขศาสตร์ \$\{item\.satSum\}",\n\s*fontSize = 11\.sp,\n\s*fontWeight = FontWeight\.Bold,\n\s*color = PurpleBadgeText)(\n)',
    r'\1,\n                            maxLines = 1,\n                            softWrap = false\2',
    content
)

content = re.sub(
    r'(text = "เงา \$\{item\.shaSum\}",\n\s*fontSize = 11\.sp,\n\s*fontWeight = FontWeight\.Bold,\n\s*color = GreenBadgeText)(\n)',
    r'\1,\n                                maxLines = 1,\n                                softWrap = false\2',
    content
)

with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
