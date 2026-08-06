package com.miracle.names

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.viewModels
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Button
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedButton
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.delay
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import com.miracle.names.ui.blessing.BlessingCardScreen
import com.miracle.names.ui.guide.MethodologyGuideSheet
import com.miracle.names.ui.results.NameResultsScreen
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.MiracleNamesTheme
import com.miracle.names.ui.wizard.SearchWizardScreen
import com.miracle.names.viewmodel.NameSearchViewModel
import com.miracle.names.viewmodel.UiState

class MainActivity : ComponentActivity() {

    private val viewModel: NameSearchViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MiracleNamesTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    MainContent(viewModel = viewModel)
                }
            }
        }
    }
}

@Composable
fun MainContent(viewModel: NameSearchViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    val selectedItem by viewModel.selectedItem.collectAsState()
    val showGuideSheet by viewModel.showGuideSheet.collectAsState()
    val connectionError by viewModel.connectionError.collectAsState()
    val isCheckingConnection by viewModel.isCheckingConnection.collectAsState()

    when {
        isCheckingConnection -> MagicalLoadingScreen()
        connectionError != null -> ErrorScreen(
            message = connectionError!!,
            onRetry = { viewModel.retryBackendConnection() }
        )
        selectedItem != null -> {
        BlessingCardScreen(item = selectedItem!!, viewModel = viewModel)
        }
        else -> {
            when (val state = uiState) {
                is UiState.Idle -> {
                    SearchWizardScreen(viewModel = viewModel)
                }
                is UiState.Loading -> {
                    MagicalLoadingScreen()
                }
                is UiState.Success -> {
                    NameResultsScreen(
                        results = state.results,
                        mode = state.mode,
                        viewModel = viewModel
                    )
                }
                is UiState.Error -> {
                    ErrorScreen(
                        message = state.message,
                        onRetry = { viewModel.searchNames() },
                        onBack = { viewModel.resetToForm() }
                    )
                }
            }
        }
    }

    // 3. Show Methodology Guide Sheet if triggered
    if (showGuideSheet) {
        MethodologyGuideSheet(onDismiss = { viewModel.toggleGuideSheet(false) })
    }
}

@Composable
fun MagicalLoadingScreen() {
    val messages = listOf(
        "กำลังเปิดม่านรายชื่อมงคล...",
        "กำลังร้อยเรียงพลังของตัวเลข...",
        "กำลังค้นหาชื่อที่เหมาะกับคุณ..."
    )
    
    var messageIndex by remember { mutableStateOf(0) }
    
    LaunchedEffect(Unit) {
        while(true) {
            delay(1200)
            messageIndex = (messageIndex + 1) % messages.size
        }
    }
    
    val infiniteTransition = rememberInfiniteTransition(label = "magical_loading")
    val rotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(3500, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "rotation"
    )
    val shimmerOffset by infiniteTransition.animateFloat(
        initialValue = -400f,
        targetValue = 1400f,
        animationSpec = infiniteRepeatable(
            animation = tween(1300, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmer"
    )
    val skeletonBase = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.7f)
    val shimmerBrush = Brush.linearGradient(
        colors = listOf(skeletonBase, GoldPrimary.copy(alpha = 0.28f), skeletonBase),
        start = Offset(shimmerOffset, 0f),
        end = Offset(shimmerOffset + 420f, 0f)
    )
    
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(horizontal = 16.dp)
    ) {
        Spacer(modifier = Modifier.height(28.dp))
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(
                imageVector = Icons.Default.AutoAwesome,
                contentDescription = null,
                tint = GoldPrimary,
                modifier = Modifier
                    .size(28.dp)
                    .rotate(rotation)
            )
            Spacer(modifier = Modifier.width(10.dp))
            Text(
                text = messages[messageIndex],
                color = GoldPrimary,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold
            )
        }

        Spacer(modifier = Modifier.height(24.dp))
        MagicalSkeletonBlock(
            modifier = Modifier.fillMaxWidth(0.58f).height(22.dp),
            brush = shimmerBrush
        )
        Spacer(modifier = Modifier.height(10.dp))
        MagicalSkeletonBlock(
            modifier = Modifier.fillMaxWidth(0.82f).height(14.dp),
            brush = shimmerBrush
        )
        Spacer(modifier = Modifier.height(20.dp))

        repeat(5) { index ->
            MagicalSkeletonNameCard(brush = shimmerBrush, isFirst = index == 0)
            Spacer(modifier = Modifier.height(10.dp))
        }
    }
}

@Composable
private fun MagicalSkeletonNameCard(brush: Brush, isFirst: Boolean) {
    val shape = RoundedCornerShape(12.dp)
    val cardColor = if (isFirst) GoldPrimary.copy(alpha = 0.08f) else MaterialTheme.colorScheme.surface

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(cardColor, shape)
            .padding(14.dp)
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            MagicalSkeletonBlock(
                modifier = Modifier.size(38.dp),
                brush = brush,
                rounded = true
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                MagicalSkeletonBlock(
                    modifier = Modifier.fillMaxWidth(if (isFirst) 0.52f else 0.42f).height(20.dp),
                    brush = brush
                )
                Spacer(modifier = Modifier.height(9.dp))
                MagicalSkeletonBlock(
                    modifier = Modifier.fillMaxWidth(0.85f).height(12.dp),
                    brush = brush
                )
            }
        }
        Spacer(modifier = Modifier.height(13.dp))
        MagicalSkeletonBlock(
            modifier = Modifier.fillMaxWidth(0.45f).height(24.dp),
            brush = brush
        )
    }
}

@Composable
private fun MagicalSkeletonBlock(
    modifier: Modifier,
    brush: Brush,
    rounded: Boolean = false
) {
    val shape = if (rounded) RoundedCornerShape(50) else RoundedCornerShape(6.dp)
    Box(modifier = modifier.background(brush, shape))
}

@Composable
fun ErrorScreen(message: String, onRetry: () -> Unit, onBack: (() -> Unit)? = null) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = Icons.Default.Warning,
                contentDescription = "Error",
                tint = MaterialTheme.colorScheme.error,
                modifier = Modifier.size(64.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "การเชื่อมต่อล้มเหลว",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurface
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(24.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                if (onBack != null) {
                    OutlinedButton(onClick = onBack) {
                        Text("กลับไปแก้ไข")
                    }
                }
                Button(onClick = onRetry) {
                    Text("ลองใหม่")
                }
            }
        }
    }
}
