package com.miracle.names

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.viewModels
import androidx.compose.material.icons.automirrored.filled.Chat
import androidx.compose.material3.BadgedBox
import androidx.compose.material3.Badge
import androidx.compose.material3.FloatingActionButton
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
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
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Switch
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
import androidx.compose.ui.zIndex
import com.miracle.names.ui.blessing.BlessingCardScreen
import com.miracle.names.ui.components.MagicalLoadingScreen
import com.miracle.names.ui.guide.MethodologyGuideSheet
import com.miracle.names.ui.results.NameResultsScreen
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.MiracleNamesTheme
import com.miracle.names.ui.wizard.SearchWizardScreen
import com.miracle.names.viewmodel.NameSearchViewModel
import com.miracle.names.viewmodel.SupportChatViewModel
import com.miracle.names.viewmodel.AuspiciousViewModel
import com.miracle.names.ui.chat.SupportChatScreen
import com.miracle.names.ui.auspicious.AuspiciousCalendarScreen
import com.miracle.names.viewmodel.UiState
import com.miracle.names.BuildConfig
import com.miracle.names.data.remote.RetrofitClient
import coil.Coil
import coil.ImageLoader

class MainActivity : ComponentActivity() {

    private val viewModel: NameSearchViewModel by viewModels()
    private val supportChatViewModel: SupportChatViewModel by viewModels()
    private val auspiciousViewModel: AuspiciousViewModel by viewModels()
    private var openChatRequested by mutableStateOf(false)

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()
        super.onCreate(savedInstanceState)
        viewModel.initializeBilling(this)
        checkChatIntent(intent)

        val imageLoader = ImageLoader.Builder(this)
            .okHttpClient(RetrofitClient.okHttpClient)
            .build()
        Coil.setImageLoader(imageLoader)

        setContent {
            MiracleNamesTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    MainContent(
                        viewModel = viewModel,
                        supportChatViewModel = supportChatViewModel,
                        auspiciousViewModel = auspiciousViewModel,
                        openChatRequested = openChatRequested,
                        onChatOpened = { openChatRequested = false }
                    )
                }
            }
        }
    }

    override fun onResume() {
        super.onResume()
        clearNotifications()
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        checkChatIntent(intent)
        clearNotifications()
    }

    private fun checkChatIntent(intent: Intent?) {
        if (intent == null) return
        val openChat = intent.getBooleanExtra("open_chat", false) ||
                intent.action == "OPEN_SUPPORT_CHAT" ||
                intent.hasExtra("conversation_id")
        if (openChat) {
            openChatRequested = true
        }
    }

    private fun clearNotifications() {
        val notificationManager = getSystemService(android.content.Context.NOTIFICATION_SERVICE) as? android.app.NotificationManager
        notificationManager?.cancelAll()
    }
}

@Composable
fun MainContent(
    viewModel: NameSearchViewModel,
    supportChatViewModel: SupportChatViewModel,
    auspiciousViewModel: AuspiciousViewModel,
    openChatRequested: Boolean = false,
    onChatOpened: () -> Unit = {}
) {
    val uiState by viewModel.uiState.collectAsState()
    val selectedItem by viewModel.selectedItem.collectAsState()
    val showGuideSheet by viewModel.showGuideSheet.collectAsState()
    val articles by viewModel.articles.collectAsState()
    val connectionError by viewModel.connectionError.collectAsState()
    val isCheckingConnection by viewModel.isCheckingConnection.collectAsState()
    val appliedResult = uiState as? UiState.Success
    var showChat by remember { mutableStateOf(false) }
    var showAuspicious by remember { mutableStateOf(false) }
    var showDebugPurchase by remember { mutableStateOf(false) }
    val chatState by supportChatViewModel.state.collectAsState()
    val mockPurchaseEnabled by viewModel.mockPurchaseEnabled.collectAsState()
    val openDebugPurchase: (() -> Unit)? = if (BuildConfig.DEBUG) {
        {
            // Emulator shortcut: the bug button enables the paid preview state.
            viewModel.setMockPurchaseEnabled(true)
            showDebugPurchase = true
        }
    } else {
        null
    }

    LaunchedEffect(openChatRequested) {
        if (openChatRequested) {
            showChat = true
            supportChatViewModel.open()
            onChatOpened()
        }
    }

    Box(Modifier.fillMaxSize()) {
    when {
        showChat -> SupportChatScreen(
            viewModel = supportChatViewModel,
            onBack = {
                supportChatViewModel.close()
                showChat = false
            }
        )
        showAuspicious -> AuspiciousCalendarScreen(
            viewModel = auspiciousViewModel,
            onBack = { showAuspicious = false }
        )
        isCheckingConnection -> MagicalLoadingScreen()
        connectionError != null -> ErrorScreen(
            message = connectionError!!,
            onRetry = { viewModel.retryBackendConnection() }
        )
        selectedItem != null && appliedResult != null -> {
            BlessingCardScreen(
                item = selectedItem!!,
                snapshot = appliedResult.snapshot,
                viewModel = viewModel,
                onOpenDebugPurchase = openDebugPurchase
            )
        }
        else -> {
            when (val state = uiState) {
                is UiState.Idle -> {
                    SearchWizardScreen(
                        viewModel = viewModel,
                        onOpenChat = {
                            supportChatViewModel.open()
                            showChat = true
                        },
                        onOpenAuspicious = { showAuspicious = true },
                        unreadCount = chatState.unreadCount,
                        onOpenDebugPurchase = openDebugPurchase
                    )
                }
                is UiState.Loading -> {
                    MagicalLoadingScreen()
                }
                is UiState.Success -> {
                    NameResultsScreen(
                        state = state,
                        viewModel = viewModel,
                        onOpenAuspicious = { showAuspicious = true },
                        onOpenDebugPurchase = openDebugPurchase
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

    if (!showChat && !isCheckingConnection) {
        FloatingActionButton(
            onClick = {
                supportChatViewModel.open()
                showChat = true
            },
            modifier = Modifier.align(Alignment.BottomEnd).padding(20.dp),
            containerColor = GoldPrimary,
            contentColor = Color(0xFF26160E)
        ) {
            BadgedBox(
                badge = {
                    if (chatState.unreadCount > 0) {
                        Badge(
                            containerColor = Color(0xFFD32F2F),
                            contentColor = Color.White
                        ) {
                            Text(
                                text = if (chatState.unreadCount > 99) "99+" else "${chatState.unreadCount}",
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                }
            ) {
                Icon(Icons.AutoMirrored.Filled.Chat, contentDescription = "ปรึกษาคุณทญา")
            }
        }
    }

    if (BuildConfig.DEBUG && showDebugPurchase) {
        AlertDialog(
            onDismissRequest = { showDebugPurchase = false },
            title = { Text("Debug purchase") },
            text = {
                Column {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Mock purchase สำเร็จ")
                        Switch(
                            checked = mockPurchaseEnabled,
                            onCheckedChange = viewModel::setMockPurchaseEnabled
                        )
                    }
                    Spacer(Modifier.height(8.dp))
                    Text(
                        if (mockPurchaseEnabled) "แอปจะแสดงชื่อเต็มโดยไม่เรียก Google Play"
                        else "แอปจะแสดงชื่อแบบล็อกตามสถานะการซื้อจริง"
                    )
                }
            },
            confirmButton = {
                Button(onClick = { showDebugPurchase = false }) { Text("ปิด") }
            }
        )
    }

    }

    // 3. Show Methodology Guide Sheet if triggered
    if (showGuideSheet) {
        MethodologyGuideSheet(
            articles = articles,
            onDismiss = { viewModel.toggleGuideSheet(false) }
        )
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
