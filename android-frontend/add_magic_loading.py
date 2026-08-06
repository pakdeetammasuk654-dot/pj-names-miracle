import re

with open("app/src/main/java/com/miracle/names/MainActivity.kt", "r", encoding="utf-8") as f:
    content = f.read()

# Add imports
imports_to_add = """import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material3.Icon
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.delay
"""

content = content.replace("import androidx.compose.runtime.getValue", imports_to_add + "import androidx.compose.runtime.getValue")

# Replace loading UI
old_loading = """            is UiState.Loading -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(color = GoldPrimary)
                }
            }"""

new_loading = """            is UiState.Loading -> {
                MagicalLoadingScreen()
            }"""

content = content.replace(old_loading, new_loading)

# Append MagicalLoadingScreen composable
magical_screen = """
@Composable
fun MagicalLoadingScreen() {
    val messages = listOf(
        "✨ กำลังรวบรวมพลังดวงดาว...",
        "🔮 กำลังผูกดวงชะตากับตัวเลข...",
        "🛡️ กำลังคัดกรองกาลกิณี...",
        "📜 กำลังถอดรหัสความหมายชื่อ...",
        "🌟 ใกล้สำเร็จแล้ว..."
    )
    
    var messageIndex by remember { mutableStateOf(0) }
    
    LaunchedEffect(Unit) {
        while(true) {
            delay(1200)
            messageIndex = (messageIndex + 1) % messages.size
        }
    }
    
    val infiniteTransition = rememberInfiniteTransition(label = "spin")
    val rotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(2000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "rotation"
    )
    
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xE6FFFFFF)),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                imageVector = Icons.Default.AutoAwesome,
                contentDescription = "Magical Star",
                tint = GoldPrimary,
                modifier = Modifier
                    .size(80.dp)
                    .rotate(rotation)
            )
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = messages[messageIndex],
                color = GoldPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold
            )
        }
    }
}
"""

content += magical_screen

with open("app/src/main/java/com/miracle/names/MainActivity.kt", "w", encoding="utf-8") as f:
    f.write(content)
