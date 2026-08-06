import re
with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "r", encoding="utf-8") as f:
    content = f.read()

# Replace the badges with PairCircles
content = re.sub(
    r'\s*// Compact Circular Badges for Sat & Sha Sums\s*Box\(\s*modifier = Modifier.*?\}\s*if \(item\.shaSum > 0\) \{\s*Box\(\s*modifier = Modifier.*?\}\s*\}',
    '''
                    // Compact Circular Badges for Sat & Sha Sums
                    if (item.satPairsDisplay.isNotEmpty()) {
                        PairCircles(pairsStr = item.satPairsDisplay, typesStr = item.satPairType)
                    }
                    if (item.shaSum > 0 && item.shaPairsDisplay.isNotEmpty()) {
                        Spacer(modifier = Modifier.width(4.dp))
                        PairCircles(pairsStr = item.shaPairsDisplay, typesStr = item.shaPairType)
                    }''',
    content,
    flags=re.DOTALL
)

# Add PairCircles composable to the end of the file
composable = """

@Composable
fun PairCircles(pairsStr: String, typesStr: String) {
    val pairs = pairsStr.split(",").map { it.trim() }.filter { it.isNotEmpty() }
    val types = typesStr.split("/").map { it.trim() }.filter { it.isNotEmpty() }

    Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
        pairs.forEachIndexed { index, pair ->
            val type = types.getOrNull(index) ?: ""
            val bgColor = when {
                type.startsWith("D10") -> Color(0xFFFFD700) // Gold
                type.startsWith("D8") -> Color(0xFFC0C0C0) // Silver
                type.startsWith("D5") -> Color(0xFFCD7F32) // Bronze
                type.startsWith("R") -> Color(0xFF87CEEB) // Sky Blue
                type.startsWith("B") -> Color(0xFFFF6B6B) // Red
                else -> Color.DarkGray
            }
            Box(
                modifier = Modifier
                    .size(24.dp)
                    .clip(androidx.compose.foundation.shape.CircleShape)
                    .background(bgColor),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = pair,
                    color = Color.Black,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}
"""

if "fun PairCircles" not in content:
    content += composable

with open("app/src/main/java/com/miracle/names/ui/results/NameResultsScreen.kt", "w", encoding="utf-8") as f:
    f.write(content)
