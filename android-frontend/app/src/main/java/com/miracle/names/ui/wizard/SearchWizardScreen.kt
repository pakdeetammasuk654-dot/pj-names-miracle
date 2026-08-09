package com.miracle.names.ui.wizard

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.animateScrollBy
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.animation.AnimatedVisibility
import coil.compose.AsyncImage
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset

import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.core.tween
import androidx.compose.animation.core.LinearOutSlowInEasing
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.RepeatMode
import androidx.compose.ui.graphics.graphicsLayer
import kotlinx.coroutines.delay
import kotlin.math.roundToInt
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Book
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.material3.Slider
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.layout
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.foundation.BorderStroke
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.GoldLight
import com.miracle.names.ui.theme.PurpleHeader
import com.miracle.names.ui.theme.PurpleDark
import com.miracle.names.ui.theme.TextSecondary
import com.miracle.names.ui.theme.CardBgDark
import com.miracle.names.ui.results.PairCircles
import com.miracle.names.ui.components.TriangularWeightSelector
import androidx.compose.foundation.shape.CircleShape
import com.miracle.names.viewmodel.NameSearchViewModel

@Composable
fun ModeSegmentedControl(
    selectedMode: String,
    onModeSelected: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f),
        border = BorderStroke(1.dp, MaterialTheme.colorScheme.outline.copy(alpha = 0.15f)),
        modifier = modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(4.dp),
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            val options = listOf(
                "standalone" to ("👤" to "เฉพาะชื่อตัว"),
                "combined" to ("📜" to "หาชื่อ + สกุล")
            )

            options.forEach { (modeKey, info) ->
                val isSelected = selectedMode == modeKey
                val (icon, label) = info

                Box(
                    modifier = Modifier
                        .weight(1f)
                        .clip(RoundedCornerShape(8.dp))
                        .then(
                            if (isSelected) {
                                Modifier.background(
                                    brush = Brush.horizontalGradient(
                                        colors = listOf(
                                            GoldPrimary.copy(alpha = 0.9f),
                                            GoldPrimary
                                        )
                                    )
                                )
                            } else {
                                Modifier.background(Color.Transparent)
                            }
                        )
                        .clickable { onModeSelected(modeKey) }
                        .padding(vertical = 10.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.Center
                    ) {
                        Text(
                            text = icon,
                            fontSize = 13.sp,
                            modifier = Modifier.padding(end = 6.dp)
                        )
                        Text(
                            text = label,
                            fontSize = 13.sp,
                            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                            color = if (isSelected) Color.Black else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SearchWizardScreen(viewModel: NameSearchViewModel) {
    val searchMode by viewModel.searchMode.collectAsState()
    val isFindSurnameMode by viewModel.isFindSurnameMode.collectAsState()
    val surname by viewModel.surname.collectAsState()
    val dayOfBirth by viewModel.dayOfBirth.collectAsState()
    val prototypeName by viewModel.prototypeName.collectAsState()
    val selectedWish by viewModel.selectedWish.collectAsState()
    val pendingWeights by viewModel.pendingRankingWeights.collectAsState()
    val celebrities by viewModel.celebrities.collectAsState()
    val decodedPairs by viewModel.decodedPairs.collectAsState()
    val decodedPairTypes by viewModel.decodedPairTypes.collectAsState()
    val decodedPairIsGood by viewModel.decodedPairIsGood.collectAsState()
    val decodedMeaning by viewModel.decodedMeaning.collectAsState()
    val isDecodedFromDb by viewModel.isDecodedFromDb.collectAsState()
    val isDecodingName by viewModel.isDecodingName.collectAsState()
    val decodeError by viewModel.decodeError.collectAsState()

    val scrollState = rememberScrollState()
    val listState = androidx.compose.foundation.lazy.rememberLazyListState()
    var isAutoScrolling by remember { mutableStateOf(true) }

    // Marquee Effect for Celebrities (using dispatchRawDelta so taps aren't blocked)
    LaunchedEffect(celebrities, isAutoScrolling) {
        if (celebrities.isNotEmpty() && isAutoScrolling) {
            while (isAutoScrolling) {
                withFrameNanos {
                    listState.dispatchRawDelta(1.2f)
                }
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        // Header Bar
        TopAppBar(
            title = {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.AutoAwesome,
                        contentDescription = null,
                        tint = GoldPrimary
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "MIRACLE NAMES",
                        color = Color.White,
                        fontWeight = FontWeight.Bold,
                        fontSize = 20.sp
                    )
                }
            },
            actions = {
                IconButton(onClick = { viewModel.toggleGuideSheet(true) }) {
                    Icon(
                        imageVector = Icons.Default.Book,
                        contentDescription = "คู่มือกระบวนการคิด",
                        tint = GoldPrimary
                    )
                }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = PurpleHeader)
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(scrollState)
                .padding(vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {

            // ----------------------------------------------------
            // Celebrity Avatars (Famous Names) - Marquee
            // ----------------------------------------------------
            if (searchMode == "standalone") {
                if (celebrities.isNotEmpty()) {
                    val avatarColors = listOf(
                    Color(0xFFE57373), Color(0xFF81C784), Color(0xFF64B5F6), Color(0xFFFFB74D), Color(0xFFBA68C8)
                )

                androidx.compose.foundation.lazy.LazyRow(
                    state = listState,
                    modifier = Modifier.fillMaxWidth(),
                    contentPadding = PaddingValues(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    items(Int.MAX_VALUE) { i ->
                        val index = i % celebrities.size
                        val celeb = celebrities[index]
                        val name = celeb.name
                        val isSelected = prototypeName == name
                        val initial = name.first().toString()
                        val bgColor = avatarColors[index % avatarColors.size]

                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier
                                .width(70.dp)
                                .padding(vertical = 4.dp)
                                .clickable {
                                    isAutoScrolling = false
                                    viewModel.prototypeName.value = name
                                    viewModel.decodeCurrentName(name)
                                }
                        ) {
                            Box(
                                modifier = Modifier
                                    .size(58.dp)
                                    .then(
                                        if (isSelected) {
                                            Modifier.background(
                                                color = bgColor.copy(alpha = 0.15f),
                                                shape = CircleShape
                                            ).border(
                                                width = 3.5.dp,
                                                color = GoldPrimary,
                                                shape = CircleShape
                                            )
                                        } else {
                                            Modifier.background(
                                                color = bgColor.copy(alpha = 0.15f),
                                                shape = CircleShape
                                            ).border(
                                                width = 2.5.dp,
                                                color = bgColor,
                                                shape = CircleShape
                                            )
                                        }
                                    ),
                                contentAlignment = Alignment.Center
                            ) {
                                if (!celeb.avatar_url.isNullOrEmpty()) {
                                    AsyncImage(
                                        model = "http://10.0.2.2:8080/avatars/${celeb.avatar_url}",
                                        contentDescription = name,
                                        contentScale = ContentScale.Crop,
                                        modifier = Modifier
                                            .size(58.dp)
                                            .clip(CircleShape)
                                    )
                                } else {
                                    Text(
                                        text = initial,
                                        color = Color.White,
                                        fontSize = 14.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.height(6.dp))
                            Text(
                                text = name.split(" ")[0],
                                color = if (isSelected) GoldPrimary else Color(0xFF4F8FE8),
                                fontSize = 14.sp,
                                fontWeight = FontWeight.W700,
                                maxLines = 1,
                                textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis
                            )
                        }
                    }
                }
            }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp)
                    .background(MaterialTheme.colorScheme.surface, RoundedCornerShape(20.dp))
                    .border(
                        width = 1.dp,
                        color = MaterialTheme.colorScheme.outlineVariant,
                        shape = RoundedCornerShape(20.dp)
                    )
                    .padding(horizontal = 12.dp, vertical = 16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                // Prototype Name input: AI Wish Query Text & Analysis
                AnimatedVisibility(
                    visible = searchMode == "standalone",
                    enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                    exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                ) {
                    Column(
                    modifier = Modifier.fillMaxWidth(),
                    verticalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                        OutlinedTextField(
                            value = prototypeName,
                            onValueChange = {
                                viewModel.prototypeName.value = it
                                viewModel.decodeCurrentName(it)
                            },
                            label = { Text("พิมพ์ชื่อต้นแบบเพื่อวิเคราะห์อ้างอิง") },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                            isError = decodeError != null,
                            supportingText = decodeError?.let { message ->
                                { Text(text = message, fontSize = 11.sp) }
                            },
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = GoldPrimary,
                                unfocusedBorderColor = Color.Gray,
                                focusedTextColor = MaterialTheme.colorScheme.onSurface,
                                unfocusedTextColor = MaterialTheme.colorScheme.onSurface
                            ),
                            trailingIcon = {
                                if (isDecodingName) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.size(18.dp),
                                        strokeWidth = 2.dp,
                                        color = GoldPrimary
                                    )
                                } else if (prototypeName.isNotEmpty()) {
                                    IconButton(onClick = {
                                        viewModel.prototypeName.value = ""
                                        viewModel.decodeCurrentName("")
                                    }) {
                                        Icon(imageVector = Icons.Default.Clear, contentDescription = "ล้างข้อความ")
                                    }
                                }
                            }
                        )

                        // ----------------------------------------------------
                        // Real-time Analysis Section (Backend API decoded pairs)
                        // ----------------------------------------------------
                        AnimatedVisibility(
                            visible = prototypeName.isNotBlank() && decodedPairs.isNotEmpty(),
                            enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                            exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                        ) {
                            val isDarkTheme = isSystemInDarkTheme()
                            Surface(
                                shape = RoundedCornerShape(12.dp),
                                color = if (isDarkTheme) Color(0xFF143324) else Color(0xFFE8F5E9),
                                border = BorderStroke(1.dp, if (isDarkTheme) Color(0xFF2E7D32) else Color(0xFFA5D6A7)),
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Column(
                                    modifier = Modifier.padding(12.dp),
                                    verticalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    if (isDecodedFromDb && decodedMeaning.isNotBlank()) {
                                        Text(
                                            text = "💡 ความหมายของชื่อต้นแบบ:",
                                            fontSize = 12.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = if (isDarkTheme) Color(0xFF81C784) else Color(0xFF2E7D32)
                                        )
                                        Text(
                                            text = decodedMeaning,
                                            fontSize = 13.sp,
                                            color = if (isDarkTheme) Color(0xFFF1F8E9) else Color(0xFF1B4332),
                                            lineHeight = 18.sp
                                        )
                                        HorizontalDivider(
                                            color = if (isDarkTheme) Color(0xFF2E7D32).copy(alpha = 0.5f) else Color(0xFFA5D6A7).copy(alpha = 0.6f),
                                            thickness = 0.5.dp
                                        )
                                    }

                                    Row(
                                        modifier = Modifier.fillMaxWidth(),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        Text(
                                            text = "คู่เลขศาสตร์ และพลังเงา:",
                                            fontSize = 12.sp,
                                            fontWeight = FontWeight.Medium,
                                            color = if (isDarkTheme) Color(0xFF81C784) else Color(0xFF2E7D32),
                                            modifier = Modifier.padding(end = 6.dp)
                                        )
                                        PairCircles(
                                            pairsStr = decodedPairs.joinToString(","),
                                            typesStr = decodedPairTypes.joinToString(","),
                                            goodFlags = decodedPairIsGood,
                                            onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                                        )
                                    }
                                }
                            }
                        }
                    }
                }


                ModeSegmentedControl(
                    selectedMode = searchMode,
                    onModeSelected = viewModel::updateSearchMode
                )

                // If combined mode, show Surname field
                AnimatedVisibility(
                    visible = searchMode == "combined",
                    enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                    exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                ) {
                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 4.dp, vertical = 0.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.End
                        ) {
                            Text(
                                text = "เปลี่ยนเป็น [หานามสกุล] + ชื่อ",
                                fontSize = 11.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.85f)
                            )
                            Spacer(modifier = Modifier.width(2.dp))
                            Box(
                                modifier = Modifier
                                    .requiredHeight(22.dp)
                                    .requiredWidth(38.dp),
                                contentAlignment = Alignment.Center
                            ) {
                                Switch(
                                    checked = isFindSurnameMode,
                                    onCheckedChange = { viewModel.isFindSurnameMode.value = it },
                                    colors = SwitchDefaults.colors(
                                        checkedThumbColor = GoldPrimary,
                                        checkedTrackColor = GoldPrimary.copy(alpha = 0.5f)
                                    ),
                                    modifier = Modifier.scale(0.55f)
                                )
                            }
                        }

                        OutlinedTextField(
                            value = surname,
                            onValueChange = {
                                viewModel.surname.value = it
                                viewModel.decodeCurrentName(it)
                            },
                            label = { Text(if (isFindSurnameMode) "กรอกชื่อตัว (เช่น ทอง)" else "กรอกนามสกุล (เช่น สมบูรณ์)") },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                            isError = decodeError != null,
                            supportingText = decodeError?.let { message ->
                                { Text(text = message, fontSize = 11.sp) }
                            },
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = GoldPrimary,
                                unfocusedBorderColor = Color.Gray,
                                focusedTextColor = MaterialTheme.colorScheme.onSurface,
                                unfocusedTextColor = MaterialTheme.colorScheme.onSurface
                            ),
                            trailingIcon = {
                                if (isDecodingName) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.size(18.dp),
                                        strokeWidth = 2.dp,
                                        color = GoldPrimary
                                    )
                                } else if (surname.isNotEmpty()) {
                                    IconButton(onClick = {
                                        viewModel.surname.value = ""
                                        viewModel.decodeCurrentName("")
                                    }) {
                                        Icon(imageVector = Icons.Default.Clear, contentDescription = "ล้างข้อความ")
                                    }
                                }
                            }
                        )

                        AnimatedVisibility(
                            visible = surname.isNotBlank() && decodedPairs.isNotEmpty(),
                            enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                            exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                        ) {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "คู่นามสกุล:",
                                    fontSize = 12.sp,
                                    color = TextSecondary,
                                    modifier = Modifier.padding(end = 6.dp)
                                )
                                PairCircles(
                                    pairsStr = decodedPairs.joinToString(","),
                                    typesStr = decodedPairTypes.joinToString(","),
                                    goodFlags = decodedPairIsGood,
                                    onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                                )
                            }
                        }
                    }
                }

                // Day of Birth Selector (Dropdown)

                val daysList = listOf(
                    "sunday" to "วันอาทิตย์",
                    "monday" to "วันจันทร์",
                    "tuesday" to "วันอังคาร",
                    "wednesday1" to "วันพุธ (กลางวัน)",
                    "wednesday2" to "วันพุธ (กลางคืน)",
                    "thursday" to "วันพฤหัสบดี",
                    "friday" to "วันศุกร์",
                    "saturday" to "วันเสาร์"
                )

                var dayDropdownExpanded by remember { mutableStateOf(false) }
                val selectedDayName = daysList.find { it.first == dayOfBirth }?.second ?: "โปรดเลือกวันเกิด..."

                ExposedDropdownMenuBox(
                    expanded = dayDropdownExpanded,
                    onExpandedChange = { dayDropdownExpanded = !dayDropdownExpanded },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    OutlinedTextField(
                        value = selectedDayName,
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("วันเกิดประจำตัว") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = dayDropdownExpanded) },
                        colors = ExposedDropdownMenuDefaults.outlinedTextFieldColors(),
                        modifier = Modifier
                            .menuAnchor()
                            .fillMaxWidth()
                    )
                    ExposedDropdownMenu(
                        expanded = dayDropdownExpanded,
                        onDismissRequest = { dayDropdownExpanded = false }
                    ) {
                        daysList.forEach { (key, name) ->
                            DropdownMenuItem(
                                text = {
                                    Text(
                                        text = name,
                                        fontWeight = if (dayOfBirth == key) FontWeight.Bold else FontWeight.Normal,
                                        color = if (dayOfBirth == key) GoldPrimary else MaterialTheme.colorScheme.onSurface
                                    )
                                },
                                onClick = {
                                    viewModel.dayOfBirth.value = key
                                    dayDropdownExpanded = false
                                }
                            )
                        }
                    }
                }


            }

            Spacer(modifier = Modifier.height(8.dp))

            // Randomized & Animated Preset Quick Selection Chips - Classic Compact Rows
            val wishPool = remember {
                listOf(
                    "🌸 เสริมดวงมงคลรอบด้าน (ทุกมิติชีวิต)",
                    "🎓 ปัญญาเฉลียวฉลาด เรียนเก่ง ความจำดี",
                    "💰 ร่ำรวยเป็นเศรษฐี มีทรัพย์สินมั่งคั่ง",
                    "👑 ผู้นำมีอำนาจบารมี ยศถาบรรดาศักดิ์",
                    "💖 เสน่ห์เมตตามหานิยม ผู้ใหญ่อุปถัมภ์",
                    "🌟 โชคลาภเงินทองไหลมาเทมา ไม่ขาดสาย",
                    "🛡️ สุขภาพแข็งแรง ปราศจากโรคภัยไข้เจ็บ",
                    "💼 การงานก้าวหน้า เจริญรุ่งเรือง เลื่อนตำแหน่ง",
                    "✨ เจรจาส่งผลสำเร็จ ผู้คนรักใคร่เอ็นดู",
                    "🔥 ดวงปังดวงเฮง ค้าขายรวยมั่งคั่ง"
                )
            }
            var currentWishList by remember {
                mutableStateOf(listOf("🌸 เสริมดวงมงคลรอบด้าน (ทุกมิติชีวิต)") + wishPool.drop(1).shuffled().take(3))
            }
            var wishAnimKey by remember { mutableIntStateOf(0) }

            val dbMeaningVirtualWish = if (decodedMeaning.isNotBlank()) "💡 $decodedMeaning" else "💡 ความหมายจากชื่อต้นแบบ"

            LaunchedEffect(selectedWish, isDecodedFromDb, decodedMeaning) {
                val pinnedItems = mutableListOf<String>()
                if (selectedWish.isNotEmpty() && wishPool.contains(selectedWish)) {
                    pinnedItems.add(selectedWish)
                }
                if (isDecodedFromDb) {
                    pinnedItems.add(dbMeaningVirtualWish)
                }

                // Instant update: Only update if pinned items aren't already at the top in the correct order
                val currentPinned = currentWishList.take(pinnedItems.size)
                if (currentPinned != pinnedItems) {
                    val remaining = wishPool.filter { it !in pinnedItems }
                    currentWishList = pinnedItems + remaining.shuffled().take(4 - pinnedItems.size)
                    wishAnimKey++
                }

                // Auto-refresh timer: shuffle remaining items every 9 seconds
                while (true) {
                    delay(9000L)
                    val remaining = wishPool.filter { it !in pinnedItems }
                    currentWishList = pinnedItems + remaining.shuffled().take(4 - pinnedItems.size)
                    wishAnimKey++
                }
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "✨ ความหมายมงคลที่ต้องการเสริม",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.85f)
                )
                IconButton(
                    onClick = {
                        val pinnedItems = mutableListOf<String>()
                        if (selectedWish.isNotEmpty() && wishPool.contains(selectedWish)) {
                            pinnedItems.add(selectedWish)
                        }
                        if (isDecodedFromDb) {
                            pinnedItems.add(dbMeaningVirtualWish)
                        }

                        val remaining = wishPool.filter { it !in pinnedItems }
                        currentWishList = pinnedItems + remaining.shuffled().take(4 - pinnedItems.size)
                        wishAnimKey++
                    },
                    modifier = Modifier.size(24.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Refresh,
                        contentDescription = "สุ่มใหม่",
                        tint = GoldPrimary,
                        modifier = Modifier.size(16.dp)
                    )
                }
            }


            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.65f)),
                shape = RoundedCornerShape(14.dp),
                border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.8f))
            ) {
                Column(modifier = Modifier.fillMaxWidth()) {
                    currentWishList.forEachIndexed { index, wish ->
                        key(wish, wishAnimKey) {
                            var visible by remember { mutableStateOf(false) }
                            LaunchedEffect(wish, wishAnimKey) {
                                visible = false
                                delay(index * 70L)
                                visible = true
                            }

                            val offsetX by animateFloatAsState(
                                targetValue = if (visible) 0f else 200f,
                                animationSpec = tween(durationMillis = 250, easing = LinearOutSlowInEasing),
                                label = "offsetX"
                            )
                            val itemAlpha by animateFloatAsState(
                                targetValue = if (visible) 1f else 0f,
                                animationSpec = tween(durationMillis = 250),
                                label = "alpha"
                            )

                            val selected = (wish == dbMeaningVirtualWish && isDecodedFromDb && selectedWish.isEmpty()) || selectedWish == wish

                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .graphicsLayer {
                                        translationX = if (selected) 0f else offsetX
                                        alpha = if (selected) 1f else itemAlpha
                                    }
                            ) {
                                val isDarkTheme = isSystemInDarkTheme()
                                val selectedBackground = if (isDarkTheme) {
                                    Brush.horizontalGradient(
                                        colors = listOf(
                                            PurpleHeader.copy(alpha = 0.96f),
                                            GoldPrimary.copy(alpha = 0.30f),
                                            PurpleHeader.copy(alpha = 0.96f)
                                        )
                                    )
                                } else {
                                    Brush.horizontalGradient(
                                        colors = listOf(
                                            Color(0xFFFFF8DE),
                                            Color(0xFFFFEAA3),
                                            Color(0xFFFFF8DE)
                                        )
                                    )
                                }
                                val selectedTextColor = if (isDarkTheme) GoldLight else PurpleHeader
                                Column(modifier = Modifier.fillMaxWidth()) {
                                    if (index > 0) {
                                        HorizontalDivider(
                                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.08f),
                                            thickness = 0.5.dp,
                                            modifier = Modifier.padding(horizontal = 10.dp)
                                        )
                                    }
                                    Row(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .clickable {
                                                if (wish == dbMeaningVirtualWish) {
                                                    if (!selected) {
                                                        viewModel.selectedWish.value = ""
                                                    }
                                                } else {
                                                    if (selected) {
                                                        viewModel.selectedWish.value = ""
                                                    } else {
                                                        viewModel.selectedWish.value = wish
                                                    }
                                                }
                                                // The LaunchedEffect will handle the reordering automatically!
                                            }
                                            .then(
                                                if (selected) {
                                                    Modifier
                                                        .background(selectedBackground)
                                                        .border(
                                                            width = 1.dp,
                                                            color = GoldPrimary.copy(alpha = if (isDarkTheme) 0.72f else 0.48f)
                                                        )
                                                } else {
                                                    Modifier.background(Color.Transparent)
                                                }
                                            )
                                            .padding(horizontal = 14.dp, vertical = 8.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        Text(
                                            text = wish,
                                            fontSize = 12.sp,
                                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
                                            color = if (selected) selectedTextColor else MaterialTheme.colorScheme.onSurface,
                                            modifier = Modifier.weight(1f),
                                            maxLines = 3,
                                            lineHeight = 18.sp
                                        )
                                        if (selected) {
                                            Text(
                                                text = "📌",
                                                color = GoldPrimary,
                                                fontWeight = FontWeight.Bold,
                                                fontSize = 12.sp
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // One ternary control replaces multiple presets and sliders (Magical Design).
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp),
                shape = RoundedCornerShape(18.dp),
                colors = CardDefaults.cardColors(containerColor = PurpleHeader),
                border = BorderStroke(
                    1.5.dp,
                    Brush.horizontalGradient(
                        listOf(
                            GoldPrimary,
                            Color(0xFFF472B6),
                            Color(0xFFA855F7),
                            GoldPrimary
                        )
                    )
                ),
                elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(
                            Brush.linearGradient(
                                colors = listOf(
                                    PurpleDark,
                                    PurpleHeader,
                                    Color(0xFF4C1D95)
                                )
                            )
                        )
                        .padding(14.dp)
                ) {
                    Column {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                text = "⚖️",
                                fontSize = 16.sp
                            )
                            Spacer(modifier = Modifier.width(6.dp))
                            Text(
                                text = "สูตรการจัดอันดับ",
                                color = Color.White,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Text(
                            text = "ลากดวงประกายไปหาด้านที่คุณให้ความสำคัญ",
                            fontSize = 11.sp,
                            color = GoldLight.copy(alpha = 0.85f),
                            modifier = Modifier.padding(top = 2.dp, bottom = 6.dp)
                        )

                        val numerologyPercent = (pendingWeights.numerology * 100).roundToInt()
                        val semanticPercent = (pendingWeights.semantic * 100).roundToInt()
                        val soundPercent = 100 - numerologyPercent - semanticPercent
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(top = 6.dp),
                            horizontalArrangement = Arrangement.spacedBy(10.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            WizardWeightPercentStack(
                                numerologyPercent = numerologyPercent,
                                semanticPercent = semanticPercent,
                                soundPercent = soundPercent,
                                modifier = Modifier.width(140.dp)
                            )
                            TriangularWeightSelector(
                                numerologyWeight = pendingWeights.numerology,
                                semanticWeight = pendingWeights.semantic,
                                soundWeight = pendingWeights.sound,
                                onWeightsChanged = viewModel::updateRankingWeights,
                                modifier = Modifier.weight(1f),
                                compact = true,
                                labelColor = Color.White.copy(alpha = 0.85f),
                                showLabels = false,
                                showCornerIcons = true
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // Action Search Button (Magical Design)
            val infiniteTransition = rememberInfiniteTransition(label = "pulse")
            val pulseScale by infiniteTransition.animateFloat(
                initialValue = 1f,
                targetValue = 1.02f,
                animationSpec = infiniteRepeatable(
                    animation = tween(1200, easing = FastOutSlowInEasing),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "scale"
            )
            val glowAlpha by infiniteTransition.animateFloat(
                initialValue = 0.3f,
                targetValue = 0.8f,
                animationSpec = infiniteRepeatable(
                    animation = tween(1200, easing = FastOutSlowInEasing),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "glow"
            )
            val shimmerShift by infiniteTransition.animateFloat(
                initialValue = -500f,
                targetValue = 1200f,
                animationSpec = infiniteRepeatable(
                    animation = tween(2200, easing = LinearOutSlowInEasing),
                    repeatMode = RepeatMode.Restart
                ),
                label = "golden_shimmer"
            )

            // A highlighted database meaning is a real selection, including in
            // name+surname mode. The UI must not ask the user to select it again.
            val hasPrototypeMeaning = isDecodedFromDb && decodedMeaning.isNotBlank()
            val hasMeaningSelection = selectedWish.isNotEmpty() || hasPrototypeMeaning

            val isFormValid = when (searchMode) {
                "standalone" -> dayOfBirth.isNotEmpty() && hasMeaningSelection
                "combined" -> dayOfBirth.isNotEmpty() && surname.isNotBlank() && hasMeaningSelection
                else -> false
            }

            val buttonText = when {
                dayOfBirth.isEmpty() -> "⚠️ โปรดเลือกวันเกิด"
                searchMode == "combined" && surname.isBlank() -> if (isFindSurnameMode) "⚠️ โปรดกรอกชื่อตัว" else "⚠️ โปรดกรอกนามสกุล"
                !hasMeaningSelection -> "⚠️ โปรดเลือกความหมาย"
                else -> "จัดอันดับชื่อดีตามตำรา"
            }

            Button(
                onClick = { viewModel.searchNames() },
                enabled = isFormValid,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp)
                    .height(64.dp)
                    .graphicsLayer {
                        if (isFormValid) {
                            scaleX = pulseScale
                            scaleY = pulseScale
                            shadowElevation = 18f * glowAlpha
                            shape = RoundedCornerShape(18.dp)
                            clip = false
                        }
                    },
                shape = RoundedCornerShape(18.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color.Transparent,
                    disabledContainerColor = Color.LightGray
                ),
                contentPadding = PaddingValues(0.dp)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            brush = Brush.linearGradient(
                                colors = if (isFormValid) listOf(
                                    GoldPrimary.copy(alpha = 0.92f),
                                    GoldPrimary,
                                    GoldLight,
                                    GoldPrimary,
                                    GoldPrimary.copy(alpha = 0.92f)
                                ) else listOf(Color.Transparent, Color.Transparent),
                            start = Offset(shimmerShift, 0f),
                            end = Offset(shimmerShift + 800f, 0f)
                        ),
                            shape = RoundedCornerShape(18.dp)
                        )
                        .border(
                            width = 1.dp,
                            color = if (isFormValid) GoldLight.copy(alpha = 0.9f) else Color.Transparent,
                            shape = RoundedCornerShape(18.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        if (isFormValid) {
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = null,
                                tint = Color.Black
                            )
                            Spacer(modifier = Modifier.width(6.dp))
                            Text(text = "🔮", fontSize = 20.sp)
                            Spacer(modifier = Modifier.width(7.dp))
                        }
                        Text(
                            text = buttonText,
                            color = if (isFormValid) Color.Black else Color.DarkGray,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = if (isFormValid) 17.sp else 15.sp,
                            letterSpacing = 0.5.sp
                        )
                        if (isFormValid) {
                            Spacer(modifier = Modifier.width(7.dp))
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = null,
                                tint = Color.Black.copy(alpha = 0.72f),
                                modifier = Modifier
                                    .size(16.dp)
                                    .graphicsLayer {
                                        scaleX = pulseScale
                                        scaleY = pulseScale
                                    }
                            )
                        }
                    }
                }
            }
            Spacer(modifier = Modifier.height(28.dp))
        }

        val selectedPairDetail by viewModel.selectedPairDetail.collectAsState()
        val isPairDetailLoading by viewModel.isPairDetailLoading.collectAsState()

        com.miracle.names.ui.components.PairDetailSheet(
            pairDetail = selectedPairDetail,
            isLoading = isPairDetailLoading,
            onDismiss = viewModel::dismissPairDetail
        )
    }
}

@Composable
fun SelectableChip(
    selected: Boolean,
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(20.dp))
            .background(if (selected) GoldPrimary else MaterialTheme.colorScheme.surface)
            .border(
                width = 1.dp,
                color = if (selected) GoldPrimary else Color.LightGray,
                shape = RoundedCornerShape(20.dp)
            )
            .clickable { onClick() }
            .padding(horizontal = 12.dp, vertical = 8.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = text,
            color = if (selected) Color.Black else MaterialTheme.colorScheme.onSurface,
            fontSize = 13.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal
        )
    }
}

@Composable
private fun WizardWeightPercentStack(
    numerologyPercent: Int,
    semanticPercent: Int,
    soundPercent: Int,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(6.dp)
    ) {
        WizardWeightPercentLine(
            icon = "⭐",
            label = "เลขศาสตร์",
            percent = numerologyPercent,
            accentColor = Color(0xFFFDE047)
        )
        WizardWeightPercentLine(
            icon = "🎯",
            label = "ความหมาย",
            percent = semanticPercent,
            accentColor = Color(0xFFF472B6)
        )
        WizardWeightPercentLine(
            icon = "🎶",
            label = "ไพเราะ",
            percent = soundPercent,
            accentColor = Color(0xFF38BDF8)
        )
    }
}

@Composable
private fun WizardWeightPercentLine(
    icon: String,
    label: String,
    percent: Int,
    accentColor: Color
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(5.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = icon,
            fontSize = 11.5.sp,
            maxLines = 1,
            modifier = Modifier.width(16.dp)
        )
        Text(
            text = label,
            color = Color.White.copy(alpha = 0.95f),
            fontSize = 11.5.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            softWrap = false,
            modifier = Modifier.weight(1f)
        )
        Text(
            text = "$percent%",
            color = accentColor,
            fontSize = 12.5.sp,
            fontWeight = FontWeight.ExtraBold,
            maxLines = 1,
            modifier = Modifier.width(36.dp)
        )
    }
}

@Composable
fun InstantDecoderUI(satSum: Int, shaSum: Int) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = Color(0xFFFDF6E3)), // Light yellowish background from image
        shape = RoundedCornerShape(8.dp),
        border = BorderStroke(2.dp, Color(0xFFFDE047))
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 16.dp, horizontal = 24.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // SatSum (เลขศาสตร์)
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .background(
                            brush = Brush.linearGradient(
                                colors = listOf(Color(0xFF4ADE80), Color(0xFF16A34A))
                            ),
                            shape = CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "$satSum",
                        color = Color(0xFF111827),
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Black
                    )
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = "เลขศาสตร์:",
                    color = Color(0xFF3D2600),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
            }

            // ShaSum (พลังเงา)
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .background(
                            brush = Brush.linearGradient(
                                colors = listOf(Color(0xFF4ADE80), Color(0xFF16A34A))
                            ),
                            shape = CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "$shaSum",
                        color = Color(0xFF111827),
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Black
                    )
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = "พลังเงา:",
                    color = Color(0xFF3D2600),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
        }
    }
}
}
