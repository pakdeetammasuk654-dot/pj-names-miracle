package com.miracle.names.ui.chat

import android.content.Context
import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.ui.res.painterResource
import androidx.compose.foundation.relocation.BringIntoViewRequester
import androidx.compose.foundation.relocation.bringIntoViewRequester
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.AttachFile
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.miracle.names.data.model.ChatMessage
import com.miracle.names.data.remote.MiracleApiConfig
import com.miracle.names.viewmodel.SupportChatViewModel
import com.miracle.names.viewmodel.ChatAvatarHelper
import coil.compose.AsyncImage
import kotlinx.coroutines.delay

import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.isImeVisible

@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class, ExperimentalLayoutApi::class)
@Composable
fun SupportChatScreen(viewModel: SupportChatViewModel, onBack: () -> Unit) {
    val state by viewModel.state.collectAsState()
    var name by remember { mutableStateOf("") }
    var text by remember { mutableStateOf("") }
    var imageBytes by remember { mutableStateOf<ByteArray?>(null) }
    var imageType by remember { mutableStateOf<String?>(null) }
    var nameFocused by remember { mutableStateOf(false) }
    var previewImageUrl by remember { mutableStateOf<String?>(null) }
    var showDeleteConfirmation by remember { mutableStateOf(false) }
    val startButtonRequester = remember { BringIntoViewRequester() }
    val scrollState = rememberScrollState()
    val listState = rememberLazyListState()
    val focusManager = LocalFocusManager.current
    val isImeVisible = WindowInsets.isImeVisible
    val context = androidx.compose.ui.platform.LocalContext.current
    val picker = rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
        if (uri != null) {
            imageBytes = context.contentResolver.openInputStream(uri)?.use { it.readBytes() }
            imageType = context.contentResolver.getType(uri) ?: "image/jpeg"
        }
    }

    DisposableEffect(Unit) {
        viewModel.open()
        onDispose {
            viewModel.close()
        }
    }

    LaunchedEffect(isImeVisible, nameFocused) {
        if (isImeVisible || nameFocused) {
            delay(120)
            scrollState.animateScrollTo(scrollState.maxValue)
            startButtonRequester.bringIntoView()
        }
    }

    LaunchedEffect(state.messages.size, state.messages.lastOrNull()?.id, isImeVisible) {
        if (state.messages.isNotEmpty()) {
            delay(60)
            listState.animateScrollToItem(state.messages.size - 1)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text("ปรึกษา อ.ทญา ว่องไว", fontWeight = FontWeight.Bold)
                        Text("Data Scientist & ผู้เชี่ยวชาญการตั้งชื่อมงคล", style = MaterialTheme.typography.labelSmall)
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "กลับ")
                    }
                },
                actions = {
                    if (state.hasConversation) {
                        TextButton(onClick = viewModel::startNewConversation) {
                            Text("เริ่มใหม่")
                        }
                        TextButton(onClick = { showDeleteConfirmation = true }, enabled = !state.isDeleting) {
                            Text(if (state.isDeleting) "กำลังลบ..." else "ลบแชต", color = MaterialTheme.colorScheme.error)
                        }
                    }
                }
            )
        }
    ) { padding ->
        Column(
            Modifier
                .fillMaxSize()
                .padding(padding)
                .imePadding()
                .background(MaterialTheme.colorScheme.background)
        ) {
            if (!state.hasConversation) {
                Column(
                    Modifier
                        .fillMaxSize()
                        .verticalScroll(scrollState)
                        .padding(horizontal = 24.dp, vertical = 18.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    Icon(Icons.Default.AutoAwesome, null, tint = Color(0xFFD4AF37), modifier = Modifier.size(40.dp))
                    Text("ปรึกษาเรื่องชื่อกับคุณทญา", style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
                    Text("ส่งคำถามและรูปประกอบ ปรึกษากับ อ.ทญา ว่องไว (Data Scientist & ผู้เชี่ยวชาญด้านศาสตร์การตั้งชื่อมงคล) ได้โดยตรง", color = Color.White.copy(alpha = 0.85f))

                    Text("เลือก avatar น่ารักของคุณ", fontWeight = FontWeight.Bold, color = Color(0xFFF3E5AB))
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        ChatAvatarHelper.options.forEach { (_, _, path) ->
                            AvatarCircle(path, state.customerAvatar, viewModel::selectAvatar)
                        }
                    }

                    OutlinedTextField(
                        value = name,
                        onValueChange = { name = it },
                        label = { Text("ชื่อที่ให้เราเรียก") },
                        placeholder = { Text("ชื่อเล่น หรือชื่อที่ใช้เรียก (ไม่บังคับ)") },
                        singleLine = true,
                        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
                        keyboardActions = KeyboardActions(onDone = { focusManager.clearFocus() }),
                        modifier = Modifier
                            .fillMaxWidth()
                            .onFocusChanged { nameFocused = it.isFocused }
                    )

                    state.error?.let {
                        Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                    }

                    Button(
                        onClick = {
                            focusManager.clearFocus()
                            viewModel.startConversation(name)
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(50.dp)
                            .bringIntoViewRequester(startButtonRequester),
                        enabled = !state.isLoading,
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Text(
                            if (state.isLoading) "กำลังเริ่มแชต..." else "เริ่มแชตกับคุณทญา →",
                            fontWeight = FontWeight.Bold
                        )
                    }

                    Spacer(Modifier.height(48.dp))
                }
            } else {
                LazyColumn(
                    state = listState,
                    modifier = Modifier.weight(1f).fillMaxWidth().padding(14.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    items(state.messages, key = { it.id }) { message ->
                        MessageBubble(
                            message = message,
                            customerAvatar = state.customerAvatar,
                            onImageClick = { fullUrl -> previewImageUrl = fullUrl }
                        )
                    }
                }

                state.error?.let {
                    Text(it, color = MaterialTheme.colorScheme.error, modifier = Modifier.padding(horizontal = 16.dp))
                }

                Column(
                    Modifier
                        .padding(12.dp)
                        .navigationBarsPadding(),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    OutlinedTextField(
                        value = text,
                        onValueChange = { text = it },
                        placeholder = { Text("พิมพ์คำถามที่อยากปรึกษา") },
                        modifier = Modifier.fillMaxWidth(),
                        minLines = 2
                    )
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        OutlinedButton(onClick = { picker.launch("image/*") }) {
                            Icon(Icons.Default.AttachFile, null)
                            Spacer(Modifier.width(4.dp))
                            Text(if (imageBytes == null) "แนบรูป" else "แนบรูปแล้ว")
                        }
                        Button(
                            onClick = {
                                viewModel.send(text, imageBytes, imageType)
                                text = ""
                                imageBytes = null
                                imageType = null
                            },
                            enabled = !state.isSending && (text.isNotBlank() || imageBytes != null),
                            modifier = Modifier.weight(1f)
                        ) {
                            Text(if (state.isSending) "กำลังส่ง..." else "ส่งข้อความ")
                        }
                    }
                }
            }
        }
    }

    if (previewImageUrl != null) {
        Dialog(
            onDismissRequest = { previewImageUrl = null },
            properties = DialogProperties(
                usePlatformDefaultWidth = false,
                dismissOnBackPress = true,
                dismissOnClickOutside = true
            )
        ) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black.copy(alpha = 0.94f))
                    .clickable { previewImageUrl = null },
                contentAlignment = Alignment.Center
            ) {
                AsyncImage(
                    model = previewImageUrl,
                    contentDescription = "รูปภาพเต็มจอ",
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(16.dp),
                    contentScale = ContentScale.Fit
                )
                IconButton(
                    onClick = { previewImageUrl = null },
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 44.dp, end = 20.dp)
                        .background(Color.Black.copy(alpha = 0.6f), CircleShape)
                ) {
                    Icon(
                        imageVector = Icons.Default.Close,
                        contentDescription = "ปิด",
                        tint = Color.White
                    )
                }
            }
        }
    }

    if (showDeleteConfirmation) {
        AlertDialog(
            onDismissRequest = { showDeleteConfirmation = false },
            title = { Text("ลบแชตและข้อมูลทั้งหมด?") },
            text = { Text("ข้อความ รูปที่แนบ และข้อมูลแจ้งเตือนของบทสนทนานี้จะถูกลบจากเซิร์ฟเวอร์ การลบย้อนกลับไม่ได้") },
            confirmButton = {
                TextButton(
                    onClick = {
                        showDeleteConfirmation = false
                        viewModel.deleteConversation()
                    },
                    enabled = !state.isDeleting
                ) { Text("ลบแชต", color = MaterialTheme.colorScheme.error) }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteConfirmation = false }) { Text("ยกเลิก") }
            }
        )
    }
}

@Composable
private fun AvatarCircle(path: String, selected: String, choose: (String) -> Unit) {
    val isSelected = selected == path
    val borderStroke = if (isSelected) androidx.compose.foundation.BorderStroke(3.dp, Color(0xFFD4AF37)) else androidx.compose.foundation.BorderStroke(1.2.dp, Color(0x33D4AF37))
    val scale = if (isSelected) 1.08f else 1.0f

    Box(
        modifier = Modifier
            .size(54.dp)
            .scale(scale)
            .clip(CircleShape)
            .border(borderStroke, CircleShape)
            .clickable { choose(path) },
        contentAlignment = Alignment.Center
    ) {
        AsyncImage(
            model = MiracleApiConfig.assetUrl(path),
            contentDescription = null,
            placeholder = painterResource(id = ChatAvatarHelper.toDrawableRes(path)),
            error = painterResource(id = ChatAvatarHelper.toDrawableRes(path)),
            contentScale = ContentScale.Crop,
            modifier = Modifier.fillMaxSize()
        )
    }
}

@Composable
private fun MessageBubble(
    message: ChatMessage,
    customerAvatar: String,
    onImageClick: (String) -> Unit = {}
) {
    val mine = message.sender == "customer"
    val avatar = if (mine) customerAvatar else "/images/taya_avatar.jpg?v=child_hd"
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = if (mine) Arrangement.End else Arrangement.Start,
        verticalAlignment = Alignment.Top
    ) {
        if (!mine) {
            AsyncImage(
                model = MiracleApiConfig.assetUrl(avatar),
                contentDescription = null,
                placeholder = painterResource(id = ChatAvatarHelper.toDrawableRes(avatar)),
                error = painterResource(id = ChatAvatarHelper.toDrawableRes(avatar)),
                modifier = Modifier.size(36.dp).clip(CircleShape).border(1.dp, Color(0xFFD4AF37), CircleShape),
                contentScale = ContentScale.Crop
            )
            Spacer(Modifier.width(8.dp))
        }
        Surface(
            color = if (mine) Color(0xFF1B4D3E) else Color(0xFF3E3314),
            shape = MaterialTheme.shapes.medium,
            tonalElevation = 2.dp,
            modifier = Modifier.widthIn(max = 280.dp)
        ) {
            Column(Modifier.padding(10.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        if (mine) "คุณ" else "อ.ทญา ว่องไว",
                        style = MaterialTheme.typography.labelSmall,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                    if (mine && message.isRead) {
                        Text(
                            text = "อ่านแล้ว",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color.White.copy(alpha = 0.55f),
                            fontWeight = FontWeight.Normal,
                            fontSize = 10.sp
                        )
                    }
                }
                if (message.body.isNotBlank()) {
                    Text(message.body, color = Color.White)
                }
                if (message.imageUrl.isNotBlank()) {
                    val fullImageUrl = MiracleApiConfig.assetUrl(message.imageUrl)
                    AsyncImage(
                        model = fullImageUrl,
                        contentDescription = "รูปที่แนบในแชต (แตะเพื่อดูเต็มจอ)",
                        modifier = Modifier
                            .fillMaxWidth()
                            .heightIn(max = 260.dp)
                            .clip(MaterialTheme.shapes.small)
                            .clickable { onImageClick(fullImageUrl) },
                        contentScale = ContentScale.Fit
                    )
                }
            }
        }
        if (mine) {
            Spacer(Modifier.width(8.dp))
            AsyncImage(
                model = MiracleApiConfig.assetUrl(avatar),
                contentDescription = null,
                placeholder = painterResource(id = ChatAvatarHelper.toDrawableRes(avatar)),
                error = painterResource(id = ChatAvatarHelper.toDrawableRes(avatar)),
                modifier = Modifier.size(36.dp).clip(CircleShape).border(1.dp, Color(0xFFD4AF37), CircleShape),
                contentScale = ContentScale.Crop
            )
        }
    }
}
