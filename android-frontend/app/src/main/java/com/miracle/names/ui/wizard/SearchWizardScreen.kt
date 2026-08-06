package com.miracle.names.ui.wizard

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.animateScrollBy
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.animation.AnimatedVisibility
import coil.compose.AsyncImage
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.draw.clip

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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.layout
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.foundation.BorderStroke
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.PurpleHeader
import com.miracle.names.ui.theme.TextSecondary
import com.miracle.names.ui.theme.CardBgDark
import com.miracle.names.ui.results.PairCircles
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
                "combined" to ("📜" to "รวมชื่อ + นามสกุล")
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
    val surname by viewModel.surname.collectAsState()
    val dayOfBirth by viewModel.dayOfBirth.collectAsState()
    val gender by viewModel.gender.collectAsState()
    val taksaTarget by viewModel.taksaTarget.collectAsState()
    val prototypeName by viewModel.prototypeName.collectAsState()
    val selectedWish by viewModel.selectedWish.collectAsState()
    val filterTwoSyllables by viewModel.filterTwoSyllables.collectAsState()
    val semanticWeight by viewModel.semanticWeight.collectAsState()
    val celebrities by viewModel.celebrities.collectAsState()
    val decodedPairs by viewModel.decodedPairs.collectAsState()
    val decodedPairTypes by viewModel.decodedPairTypes.collectAsState()
    val decodedMeaning by viewModel.decodedMeaning.collectAsState()
    val isDecodedFromDb by viewModel.isDecodedFromDb.collectAsState()

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
        SmallTopAppBar(
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
            colors = TopAppBarDefaults.smallTopAppBarColors(containerColor = PurpleHeader)
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
                    Text(
                    text = "💡 ลองเลือกชื่อคนดัง (Celebrity Avatars):",
                    fontSize = 12.sp,
                    color = TextSecondary,
                    modifier = Modifier.padding(horizontal = 16.dp)
                )
                
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
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp),
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
                            label = { Text("กรอกชื่อต้นแบบที่ต้องการค้นหา") },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = GoldPrimary,
                                unfocusedBorderColor = Color.Gray,
                                focusedTextColor = MaterialTheme.colorScheme.onSurface,
                                unfocusedTextColor = MaterialTheme.colorScheme.onSurface
                            ),
                            trailingIcon = {
                                if (prototypeName.isNotEmpty()) {
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
                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                AnimatedVisibility(
                                    visible = isDecodedFromDb && decodedMeaning.isNotBlank(),
                                    enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                                    exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                                ) {
                                    // Show DB meaning (read-only)
                                    Surface(
                                        shape = RoundedCornerShape(8.dp),
                                        color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
                                        border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.3f)),
                                        modifier = Modifier.fillMaxWidth()
                                    ) {
                                        Column(modifier = Modifier.padding(10.dp)) {
                                            Text(
                                                text = "💡 ความหมายของชื่อต้นแบบ:",
                                                fontSize = 12.sp,
                                                fontWeight = FontWeight.Bold,
                                                color = GoldPrimary
                                            )
                                            Spacer(modifier = Modifier.height(2.dp))
                                            Text(
                                                text = decodedMeaning,
                                                fontSize = 13.sp,
                                                color = MaterialTheme.colorScheme.onSurface,
                                                lineHeight = 18.sp
                                            )
                                        }
                                    }
                                }
    
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Text(
                                        text = "คู่เลข:",
                                        fontSize = 12.sp,
                                        color = TextSecondary,
                                        modifier = Modifier.padding(end = 6.dp)
                                    )
                                    PairCircles(
                                        pairsStr = decodedPairs.joinToString(","),
                                        typesStr = decodedPairTypes.joinToString(",")
                                    )
                                }
                            }
                        }
                    }
                }


                ModeSegmentedControl(
                    selectedMode = searchMode,
                    onModeSelected = { viewModel.searchMode.value = it }
                )

                // If combined mode, show Surname field
                AnimatedVisibility(
                    visible = searchMode == "combined",
                    enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                    exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                ) {
                    OutlinedTextField(
                        value = surname,
                        onValueChange = { viewModel.surname.value = it },
                        label = { Text("กรอกนามสกุล (เช่น สมบูรณ์)") },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true
                    )
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
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.25f)),
                shape = RoundedCornerShape(0.dp),
                border = BorderStroke(1.dp, MaterialTheme.colorScheme.outline.copy(alpha = 0.12f))
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

                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .graphicsLayer {
                                        translationX = offsetX
                                        alpha = itemAlpha
                                    }
                            ) {
                                val selected = (wish == dbMeaningVirtualWish && isDecodedFromDb && selectedWish.isEmpty()) || selectedWish == wish
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
                                            .background(if (selected) GoldPrimary.copy(alpha = 0.2f) else Color.Transparent)
                                            .padding(horizontal = 10.dp, vertical = 6.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        Text(
                                            text = wish,
                                            fontSize = 12.sp,
                                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
                                            color = if (selected) GoldPrimary else MaterialTheme.colorScheme.onSurface,
                                            modifier = Modifier.weight(1f),
                                            maxLines = 1
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

            // Slider for Weight in a clean Card below
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)),
                shape = RoundedCornerShape(12.dp)
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Text(
                        text = "⚖️ น้ำหนักการจัดอันดับ (Ranking Weight)",
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(4.dp))

                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(text = "เน้นเลขศาสตร์", fontSize = 12.sp, color = TextSecondary)
                        Text(text = "สมดุล", fontSize = 12.sp, color = TextSecondary)
                        Text(text = "เน้นความหมาย", fontSize = 12.sp, color = TextSecondary)
                    }
                    Slider(
                        value = semanticWeight,
                        onValueChange = { viewModel.semanticWeight.value = it },
                        valueRange = 0f..1f,
                        steps = 9,
                        modifier = Modifier.fillMaxWidth()
                    )
                    val weightPercent = (semanticWeight * 100).toInt()
                    Text(
                        text = "ความหมาย: ${weightPercent}%, เลขศาสตร์: ${100 - weightPercent}%",
                        fontSize = 12.sp,
                        color = GoldPrimary,
                        modifier = Modifier.align(Alignment.CenterHorizontally),
                        fontWeight = FontWeight.Bold
                    )
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

            val hasPrototypeMeaning = searchMode == "standalone" && isDecodedFromDb

            val isFormValid = when (searchMode) {
                "standalone" -> dayOfBirth.isNotEmpty() && (selectedWish.isNotEmpty() || hasPrototypeMeaning)
                "combined" -> dayOfBirth.isNotEmpty() && surname.isNotEmpty() && selectedWish.isNotEmpty()
                else -> false
            }

            val buttonText = when {
                dayOfBirth.isEmpty() -> "⚠️ โปรดเลือกวันเกิด"
                searchMode == "combined" && surname.isEmpty() -> "⚠️ โปรดกรอกนามสกุล"
                selectedWish.isEmpty() && !hasPrototypeMeaning -> "⚠️ โปรดเลือกความหมาย"
                else -> "🔮 จัดลำดับชื่อตามตำราที่ดีที่สุด"
            }

            Button(
                onClick = { viewModel.searchNames() },
                enabled = isFormValid,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp)
                    .height(58.dp)
                    .graphicsLayer {
                        if (isFormValid) {
                            scaleX = pulseScale
                            scaleY = pulseScale
                            shadowElevation = 10f * glowAlpha
                            shape = RoundedCornerShape(16.dp)
                            clip = false
                        }
                    },
                shape = RoundedCornerShape(16.dp),
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
                            brush = Brush.horizontalGradient(
                                colors = if (isFormValid) listOf(
                                    GoldPrimary.copy(alpha = 0.9f),
                                    GoldPrimary,
                                    GoldPrimary.copy(alpha = 0.9f)
                                ) else listOf(Color.Transparent, Color.Transparent)
                            ),
                            shape = RoundedCornerShape(16.dp)
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
                            Spacer(modifier = Modifier.width(8.dp))
                        }
                        Text(
                            text = buttonText,
                            color = if (isFormValid) Color.Black else Color.DarkGray,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = if (isFormValid) 17.sp else 15.sp,
                            letterSpacing = 0.5.sp
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.height(28.dp))
            }
        }
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
