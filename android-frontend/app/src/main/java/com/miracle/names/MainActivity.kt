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
import com.miracle.names.ui.components.MagicalLoadingScreen
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
    val appliedResult = uiState as? UiState.Success

    when {
        isCheckingConnection -> MagicalLoadingScreen()
        connectionError != null -> ErrorScreen(
            message = connectionError!!,
            onRetry = { viewModel.retryBackendConnection() }
        )
        selectedItem != null && appliedResult != null -> {
            BlessingCardScreen(
                item = selectedItem!!,
                snapshot = appliedResult.snapshot,
                viewModel = viewModel
            )
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
                        state = state,
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
