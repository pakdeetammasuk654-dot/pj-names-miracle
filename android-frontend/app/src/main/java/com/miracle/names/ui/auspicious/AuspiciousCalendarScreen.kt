package com.miracle.names.ui.auspicious

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.ui.graphics.Brush
import androidx.compose.foundation.clickable
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.relocation.BringIntoViewRequester
import androidx.compose.foundation.relocation.bringIntoViewRequester
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import kotlinx.coroutines.delay
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CalendarMonth
import androidx.compose.material.icons.filled.ChevronLeft
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material.icons.filled.ErrorOutline
import androidx.compose.material.icons.filled.ExpandLess
import androidx.compose.material.icons.filled.ExpandMore
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Palette
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TimePicker
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.material3.rememberTimePickerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.withFrameNanos
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.scale
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.AuspiciousActivity
import com.miracle.names.data.model.AuspiciousDateResult
import com.miracle.names.data.model.OutfitColor
import com.miracle.names.data.model.OutfitColorSet
import com.miracle.names.data.model.AuspiciousSearchRequest
import com.miracle.names.data.model.AuspiciousSearchResponse
import com.miracle.names.ui.theme.CardBgDark
import com.miracle.names.ui.theme.DarkColorScheme
import com.miracle.names.ui.theme.GoldLight
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.GreenHeader
import com.miracle.names.ui.theme.PromptFontFamily
import com.miracle.names.viewmodel.AuspiciousViewModel
import java.time.LocalDate
import java.time.YearMonth

private val emeraldDeep = Color(0xFF140B07)
private val emeraldCard = Color(0xFF160C08)
private val lvBase = Color(0xFF140B07)
private val lvCard = Color(0xFF160C08)
private val lvInner = Color(0xFF1F120B)
private val lvAccent = Color(0xFF26160E)
private val warmText = Color(0xFFF7EDC9)
private val mutedText = Color(0xFFD6C8BE)
private val thaiMonths = listOf("มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม")

@OptIn(ExperimentalFoundationApi::class, ExperimentalMaterial3Api::class)
@Composable
fun AuspiciousCalendarScreen(
    viewModel: AuspiciousViewModel,
    onBack: () -> Unit
) {
    val state by viewModel.state.collectAsState()
    val today = remember { LocalDate.now() }
    var startDate by rememberSaveable { mutableStateOf(today.toString()) }
    var endDate by rememberSaveable { mutableStateOf(today.plusDays(90).toString()) }
    val birthDate = state.birthDate
    val birthTime = state.birthTime
    var preferWeekend by rememberSaveable { mutableStateOf(false) }
    var avoidWanPhra by rememberSaveable { mutableStateOf(false) }
    var monkCeremony by rememberSaveable { mutableStateOf(false) }
    val birthCardRequester = remember { BringIntoViewRequester() }
    val timeFieldRequester = remember { BringIntoViewRequester() }
    val listState = rememberLazyListState()
    var formError by rememberSaveable { mutableStateOf<String?>(null) }
    var showResultsSheet by rememberSaveable { mutableStateOf(false) }
    val missingFormMessage = when {
        state.selectedActivityCode.isBlank() -> "⚠️ โปรดเลือกกิจกรรม"
        birthDate.isBlank() -> "⚠️ โปรดกรอกวันเกิด"
        birthTime.isBlank() -> "⚠️ โปรดกรอกเวลาเกิด"
        else -> null
    }
    val canSearch = missingFormMessage == null && !state.isSearching && !state.isLoadingActivities
    val pulseTransition = rememberInfiniteTransition(label = "auspiciousSearchPulse")
    val searchPulseScale by pulseTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.025f,
        animationSpec = infiniteRepeatable(
            animation = tween(durationMillis = 900),
            repeatMode = RepeatMode.Reverse
        ),
        label = "auspiciousSearchPulseScale"
    )

    LaunchedEffect(state.selectedActivityCode) {
        if (state.selectedActivityCode == "ordination") monkCeremony = true
    }
    LaunchedEffect(state.result) {
        if (state.result != null) showResultsSheet = true
    }
    LaunchedEffect(birthDate, birthTime) {
        if (birthDate.isBlank() || birthTime.isBlank()) {
            viewModel.clearTaksa()
            return@LaunchedEffect
        }
        viewModel.loadTaksa(
            birthDate = birthDate,
            birthTime = birthTime
        )
    }
    LaunchedEffect(birthTime) {
        if (birthTime.isNotBlank()) {
            delay(150)
            timeFieldRequester.bringIntoView()
        }
    }

    MaterialTheme(colorScheme = DarkColorScheme) {
        Scaffold(
            containerColor = emeraldDeep,
            topBar = {
                TopAppBar(
                    title = {
                        Column {
                            Text("ดูฤกษ์ยาม", color = Color.White, fontWeight = FontWeight.Bold)
                            Text("ตามหลักตำราอีสานและทักษาจร", color = GoldLight, fontSize = 11.sp)
                        }
                    },
                    navigationIcon = {
                        IconButton(onClick = onBack) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, "กลับ", tint = Color.White)
                        }
                    },
                    colors = TopAppBarDefaults.topAppBarColors(containerColor = GreenHeader)
                )
            }
        ) { innerPadding ->
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding),
                contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 16.dp, bottom = 40.dp),
                verticalArrangement = Arrangement.spacedBy(14.dp)
            ) {
                if (state.isLoadingActivities) {
                    item(key = "loading_activities") { LoadingBlock("กำลังโหลดรายการกิจกรรม…") }
                } else {
                    item(key = "step1_heading") {
                        StepHeading(number = "1", title = "คุณกำลังจะทำอะไร?", subtitle = "เลือกหมวด แล้วเลือกเหตุการณ์สำคัญ")
                    }
                    item(key = "activity_picker") {
                        ActivityPicker(
                            activities = state.activities,
                            selectedCode = state.selectedActivityCode,
                            onSelected = { viewModel.selectActivity(it) }
                        )
                    }
                }

                item(key = "step2_heading") {
                    StepHeading(number = "2", title = "ช่วงวันที่สะดวก", subtitle = "ระบบจะเปรียบเทียบได้สูงสุด 366 วัน")
                }
                item(key = "date_range_card") {
                    CardBlock {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            Box(Modifier.weight(1f)) {
                                val end = runCatching { LocalDate.parse(endDate) }.getOrNull()
                                DateField(
                                    label = "วันเริ่มต้น",
                                    value = startDate,
                                    minimumDate = end?.minusDays(365),
                                    maximumDate = end
                                ) { startDate = it; formError = null }
                            }
                            Box(Modifier.weight(1f)) {
                                val start = runCatching { LocalDate.parse(startDate) }.getOrNull()
                                DateField(
                                    label = "วันสิ้นสุด",
                                    value = endDate,
                                    minimumDate = start,
                                    maximumDate = start?.plusDays(365)
                                ) { endDate = it; formError = null }
                            }
                        }
                    }
                }

                item(key = "step3_heading") {
                    StepHeading(number = "3", title = "ปรับให้เหมาะกับคุณ", subtitle = "วันเกิดช่วยให้ระบบผสานทักษาจรเฉพาะบุคคล")
                }
                item(key = "birth_card") {
                    CardBlock {
                        BirthDateField(
                            value = birthDate,
                            onSelected = viewModel::updateBirthDate
                        )
                        if (birthDate.isBlank()) {
                            Text(
                                text = "แตะช่องวันเกิด แล้วเลือกวันที่จากปฏิทิน พ.ศ.",
                                color = mutedText,
                                fontSize = 11.sp,
                                lineHeight = 16.sp,
                                modifier = Modifier.padding(top = 6.dp, start = 2.dp)
                            )
                        }
                        Spacer(Modifier.height(10.dp))
                        Box(modifier = Modifier.bringIntoViewRequester(timeFieldRequester)) {
                            TimeField(value = birthTime, onSelected = viewModel::updateBirthTime)
                        }
                        if (birthTime.isBlank()) {
                            Text(
                                text = "จำเป็นต้องระบุเวลาเกิด เพื่อคำนวณทักษาจรโดยใช้เวลาเปลี่ยนวัน 06:00 น.",
                                color = GoldLight,
                                fontSize = 11.sp,
                                modifier = Modifier.padding(top = 4.dp, start = 2.dp)
                            )
                        }
                    }
                }
                if (state.isLoadingTaksa) {
                    item(key = "loading_taksa") {
                        CardBlock {
                            Row(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp),
                                horizontalArrangement = Arrangement.Center,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                CircularProgressIndicator(modifier = Modifier.size(20.dp), color = GoldPrimary, strokeWidth = 2.dp)
                                Spacer(Modifier.width(10.dp))
                                Text("กำลังคำนวณทักษาจร…", color = mutedText, fontSize = 13.sp)
                            }
                        }
                    }
                } else if (state.taksa != null) {
                    item(key = "taksa_wheel") {
                        TaksaWheelPanel(result = state.taksa!!)
                    }
                } else if (state.taksaErrorMessage != null) {
                    item(key = "taksa_error") {
                        CardBlock {
                            Column(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
                                horizontalAlignment = Alignment.CenterHorizontally
                            ) {
                                Text(state.taksaErrorMessage ?: "", color = Color(0xFFFFB4AB), fontSize = 13.sp)
                                TextButton(onClick = {
                                    viewModel.loadTaksa(
                                        birthDate = birthDate,
                                        birthTime = birthTime
                                    )
                                }) {
                                    Text("ลองคำนวณใหม่", color = GoldPrimary)
                                }
                            }
                        }
                    }
                }

                item {
                    StepHeading(number = "4", title = "ความสะดวกและรูปแบบพิธี", subtitle = "ตัวเลือกของคุณจะแยกจากกฎตำราอย่างชัดเจน")
                }
                item {
                    CardBlock {
                        PreferenceSwitch("เน้นวันเสาร์–อาทิตย์", "เพิ่มน้ำหนักด้านความสะดวก", preferWeekend) { preferWeekend = it }
                        PreferenceSwitch("หลีกเลี่ยงวันพระ", "ตัดวันพระออกจากผลลัพธ์", avoidWanPhra) { avoidWanPhra = it }
                        PreferenceSwitch("มีพิธีสงฆ์", "ตรวจข้อห้ามที่เกี่ยวกับพิธีสงฆ์", monkCeremony) { monkCeremony = it }
                    }
                }

                val displayedError = formError ?: state.errorMessage
                if (displayedError != null) {
                    item {
                        ErrorCard(
                            message = displayedError,
                            onRetry = if (state.activities.isEmpty()) viewModel::loadActivities else null,
                            onDismiss = {
                                formError = null
                                viewModel.clearError()
                            }
                        )
                    }
                }

                item {
                    Button(
                        onClick = {
                            val start = runCatching { LocalDate.parse(startDate) }.getOrNull()
                            val end = runCatching { LocalDate.parse(endDate) }.getOrNull()
                            formError = when {
                                state.selectedActivityCode.isBlank() -> "กรุณาเลือกกิจกรรมที่ต้องการหาฤกษ์"
                                birthDate.isBlank() -> "กรุณาระบุวันเกิดเพื่อคำนวณทักษาจร"
                                birthTime.isBlank() -> "กรุณาระบุเวลาเกิด เพื่อคำนวณทักษาจรโดยใช้เวลาเปลี่ยนวัน 06:00 น."
                                start == null || end == null -> "รูปแบบวันที่ไม่ถูกต้อง"
                                end.isBefore(start) -> "วันที่สิ้นสุดต้องไม่มาก่อนวันที่เริ่มต้น"
                                end.toEpochDay() - start.toEpochDay() + 1 > 366 -> "เลือกช่วงเวลาได้ไม่เกิน 366 วัน"
                                else -> null
                            }
                            if (formError == null) {
                                viewModel.search(
                                    AuspiciousSearchRequest(
                                        activityCode = state.selectedActivityCode,
                                        startDate = startDate,
                                        endDate = endDate,
                                        birthDate = birthDate,
                                        birthTime = birthTime,
                                        preferWeekend = preferWeekend,
                                        avoidWanPhra = avoidWanPhra,
                                        monkCeremony = monkCeremony
                                    )
                                )
                            }
                        },
                        enabled = canSearch,
                        modifier = Modifier
                            .fillMaxWidth()
                            .scale(if (canSearch) searchPulseScale else 1f)
                            .shadow(
                                elevation = if (canSearch) 12.dp else 0.dp,
                                shape = RoundedCornerShape(18.dp),
                                spotColor = GoldPrimary.copy(alpha = .65f)
                            )
                            .height(58.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = GoldPrimary,
                            contentColor = Color(0xFF140B07),
                            disabledContainerColor = if (state.isSearching) GoldPrimary.copy(alpha = .72f) else Color.LightGray,
                            disabledContentColor = if (state.isSearching) Color(0xFF140B07) else Color.DarkGray
                        ),
                        shape = RoundedCornerShape(18.dp)
                    ) {
                        if (state.isSearching) {
                            CircularProgressIndicator(Modifier.size(20.dp), color = Color.DarkGray, strokeWidth = 2.dp)
                            Spacer(Modifier.width(10.dp))
                            Text("กำลังเปรียบเทียบทุกวัน…", color = Color.DarkGray, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                        } else if (missingFormMessage != null) {
                            Text(
                                text = missingFormMessage,
                                color = Color.DarkGray,
                                fontWeight = FontWeight.ExtraBold,
                                fontSize = 15.sp,
                                letterSpacing = 0.5.sp
                            )
                        } else {
                            Icon(Icons.Default.CalendarMonth, null, tint = Color(0xFF140B07))
                            Spacer(Modifier.width(8.dp))
                            Text("ค้นหาวันที่เหมาะที่สุด", color = Color(0xFF140B07), fontWeight = FontWeight.ExtraBold, fontSize = 17.sp)
                        }
                    }
                }

            }
        }
        val result = state.result
        if (showResultsSheet && result != null) {
            ModalBottomSheet(
                onDismissRequest = { showResultsSheet = false },
                sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
                containerColor = emeraldCard
            ) {
                LazyColumn(
                    modifier = Modifier.fillMaxWidth(),
                    contentPadding = PaddingValues(start = 14.dp, end = 14.dp, bottom = 36.dp),
                    verticalArrangement = Arrangement.spacedBy(14.dp)
                ) {
                    item { ResultSummary(result) }
                    if (result.recommendations.isEmpty()) {
                        item { CardBlock {
                            Text("ไม่พบวันที่ผ่านทุกเงื่อนไขในช่วงนี้", color = warmText, fontWeight = FontWeight.Bold)
                            Text("ลองขยายช่วงวันที่หรือปรับตัวเลือก", color = mutedText)
                        } }
                    } else {
                        items(result.recommendations, key = { it.date }) { day ->
                            RecommendationCard(day, result.recommendations.indexOf(day))
                        }
                    }
                    item { MethodCard(result) }
                }
            }
        }
    }
}

@Composable
private fun StepHeading(number: String, title: String, subtitle: String) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(38.dp).background(GoldPrimary, CircleShape), contentAlignment = Alignment.Center) {
            Text(number, color = Color(0xFF140B07), fontWeight = FontWeight.ExtraBold)
        }
        Spacer(Modifier.width(10.dp))
        Column {
            Text(title, color = GoldLight, fontSize = 19.sp, fontWeight = FontWeight.Bold, fontFamily = PromptFontFamily)
            Text(subtitle, color = mutedText, fontSize = 13.sp)
        }
    }
}

@Composable
private fun LoadingBlock(label: String) {
    Box(Modifier.fillMaxWidth().height(120.dp), contentAlignment = Alignment.Center) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            CircularProgressIndicator(Modifier.size(24.dp), color = GoldPrimary, strokeWidth = 2.dp)
            Spacer(Modifier.width(10.dp))
            Text(label, color = mutedText)
        }
    }
}

@Composable
private fun ActivityPicker(
    activities: List<AuspiciousActivity>,
    selectedCode: String,
    onSelected: (String) -> Unit
) {
    var expanded by remember { mutableStateOf(false) }
    val selected = activities.firstOrNull { it.code == selectedCode } ?: activities.firstOrNull()
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = lvInner),
        border = BorderStroke(1.dp, GoldPrimary.copy(alpha = .45f)),
        shape = RoundedCornerShape(18.dp)
    ) {
        Box {
            OutlinedButton(
                onClick = { expanded = true },
                modifier = Modifier.fillMaxWidth().padding(8.dp),
                shape = RoundedCornerShape(12.dp),
                border = BorderStroke(0.dp, Color.Transparent),
                contentPadding = PaddingValues(horizontal = 10.dp, vertical = 7.dp)
            ) {
                Text(selected?.icon ?: "✦", fontSize = 24.sp)
                Spacer(Modifier.width(10.dp))
                Column(Modifier.weight(1f), horizontalAlignment = Alignment.Start) {
                    Text("กิจกรรมที่เลือก", color = mutedText, fontSize = 11.sp)
                    Text(
                        selected?.nameTh ?: "เลือกกิจกรรม",
                        color = GoldLight,
                        fontWeight = FontWeight.Bold,
                        fontSize = 16.sp,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis
                    )
                }
                Icon(Icons.Default.ArrowDropDown, "เปิดรายการกิจกรรม", tint = GoldPrimary, modifier = Modifier.size(28.dp))
            }
            DropdownMenu(
                expanded = expanded,
                onDismissRequest = { expanded = false },
                modifier = Modifier.background(lvAccent)
            ) {
                activities.forEach { activity ->
                    DropdownMenuItem(
                        text = {
                            Column {
                                Text("${activity.icon} ${activity.nameTh}", color = GoldLight, fontWeight = FontWeight.Bold, maxLines = 2)
                            }
                        },
                        onClick = {
                            expanded = false
                            onSelected(activity.code)
                        },
                        leadingIcon = if (activity.code == selectedCode) {
                            { Icon(Icons.Default.CheckCircle, null, tint = GoldPrimary) }
                        } else null
                    )
                }
            }
        }
    }
}

@Composable
private fun CardBlock(
    modifier: Modifier = Modifier,
    content: @Composable ColumnScope.() -> Unit
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = CardBgDark),
        border = BorderStroke(1.dp, GoldPrimary.copy(alpha = .22f)),
        shape = RoundedCornerShape(18.dp)
    ) {
        Column(Modifier.padding(16.dp), content = content)
    }
}

@Composable
private fun PreferenceSwitch(title: String, subtitle: String, checked: Boolean, onCheckedChange: (Boolean) -> Unit) {
    Row(Modifier.fillMaxWidth().padding(vertical = 5.dp), verticalAlignment = Alignment.CenterVertically) {
        Column(Modifier.weight(1f)) {
            Text(title, color = warmText, fontWeight = FontWeight.Medium)
            Text(subtitle, color = mutedText, fontSize = 12.sp)
        }
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            colors = SwitchDefaults.colors(checkedThumbColor = Color(0xFF140B07), checkedTrackColor = GoldPrimary)
        )
    }
}

@Composable
private fun ErrorCard(message: String, onRetry: (() -> Unit)?, onDismiss: () -> Unit) {
    Card(colors = CardDefaults.cardColors(containerColor = Color(0xFF4A201C)), shape = RoundedCornerShape(14.dp)) {
        Row(Modifier.fillMaxWidth().padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Icon(Icons.Default.ErrorOutline, null, tint = Color(0xFFFFB4AB))
            Spacer(Modifier.width(10.dp))
            Text(message, color = Color(0xFFFFDAD6), modifier = Modifier.weight(1f))
            onRetry?.let {
                TextButton(onClick = it) { Text("ลองใหม่", color = GoldLight) }
            }
            IconButton(onClick = onDismiss) { Icon(Icons.Default.Close, "ปิด", tint = Color(0xFFFFDAD6)) }
        }
    }
}

@Composable
private fun ResultSummary(result: AuspiciousSearchResponse) {
    CardBlock {
        Text("${result.recommendations.size} วันที่ดีที่สุด", color = GoldLight, fontSize = 22.sp, fontWeight = FontWeight.Bold)
        Text("สำหรับ ${result.activity.nameTh}", color = warmText, fontWeight = FontWeight.SemiBold)
        Spacer(Modifier.height(6.dp))
        Text("ตรวจ ${result.evaluatedDays} วัน · ตัดวันที่ไม่ผ่าน ${result.excludedDays} วัน", color = mutedText)
    }
}

@Composable
private fun RecommendationCard(day: AuspiciousDateResult, index: Int) {
    val rankLabels = listOf("แนะนำที่สุด", "เหมาะมาก", "ทางเลือกที่ดี")
    var expanded by rememberSaveable(day.date) { mutableStateOf(false) }
    Card(
        colors = CardDefaults.cardColors(containerColor = emeraldCard),
        border = BorderStroke(if (index == 0) 2.dp else 1.dp, GoldPrimary.copy(alpha = if (index == 0) .9f else .35f)),
        shape = RoundedCornerShape(20.dp)
    ) {
        Column(Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.Top) {
                Box(Modifier.size(42.dp).background(GoldPrimary, CircleShape), contentAlignment = Alignment.Center) {
                    Text("${index + 1}", color = Color(0xFF140B07), fontWeight = FontWeight.ExtraBold, fontSize = 18.sp)
                }
                Spacer(Modifier.width(12.dp))
                Column(Modifier.weight(1f)) {
                    Text(rankLabels.getOrElse(index) { "วันที่แนะนำ" }, color = GoldPrimary, fontSize = 12.sp, fontWeight = FontWeight.Bold)
                    Text(day.weekday, color = GoldLight, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                    Text(day.thaiDate, color = warmText)
                    Text(day.lunarLabel + if (day.isWanPhra) " · วันพระ" else "", color = mutedText, fontSize = 13.sp)
                    if (day.importantDay.isNotBlank()) {
                        Text("✦ ${day.importantDay}", color = GoldPrimary, fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
                    }
                }
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("${day.score}", color = GoldLight, fontSize = 28.sp, fontWeight = FontWeight.ExtraBold)
                    Text("คะแนน", color = mutedText, fontSize = 11.sp)
                }
            }
            Spacer(Modifier.height(12.dp))
            HorizontalDivider(color = GoldPrimary.copy(alpha = .18f))
            Spacer(Modifier.height(8.dp))
            Text(day.rating, color = warmText, fontWeight = FontWeight.Bold)
            day.reasons.take(if (expanded) day.reasons.size else 3).forEach { reason ->
                val color = when (reason.impact) {
                    "caution" -> Color(0xFFFFD59A)
                    "practical" -> Color(0xFFB8D7FF)
                    else -> Color(0xFFFFDF85)
                }
                Column(Modifier.padding(top = 8.dp)) {
                    Text((if (reason.impact == "caution") "⚠ " else "✓ ") + reason.name, color = color, fontWeight = FontWeight.SemiBold)
                    if (expanded) {
                        Text(reason.meaning.ifBlank { reason.explanation }, color = mutedText, fontSize = 13.sp, lineHeight = 19.sp)
                    }
                }
            }
            if (day.reasons.isNotEmpty()) {
                TextButton(onClick = { expanded = !expanded }, modifier = Modifier.align(Alignment.End)) {
                    Text(if (expanded) "ซ่อนความหมาย" else "ดูความหมายของฤกษ์", color = GoldPrimary)
                    Icon(if (expanded) Icons.Default.ExpandLess else Icons.Default.ExpandMore, null, tint = GoldPrimary)
                }
            }
            day.outfitAdvice?.let { advice ->
                if (advice.auspicious.isNotEmpty()) {
                    Spacer(Modifier.height(8.dp))
                    HorizontalDivider(color = GoldPrimary.copy(alpha = .18f))
                    Spacer(Modifier.height(8.dp))
                    OutfitColorSection(
                        title = "สีที่เสริม",
                        sets = advice.auspicious,
                        positive = true
                    )
                    if (advice.inauspicious.isNotEmpty()) {
                        Spacer(Modifier.height(8.dp))
                        OutfitColorSection(
                            title = "สีที่ควรเลี่ยง",
                            sets = advice.inauspicious,
                            positive = false
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun OutfitColorSection(
    title: String,
    sets: List<OutfitColorSet>,
    positive: Boolean
) {
    if (sets.isEmpty()) return
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(6.dp)
    ) {
        Text(
            text = title,
            color = if (positive) warmText else mutedText,
            fontSize = 12.sp,
            fontWeight = FontWeight.Bold
        )
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceAround,
            verticalAlignment = Alignment.Top
        ) {
            sets.forEach { set ->
                val displayLabel = when (set.context) {
                    "appointment" -> if (positive) "เสริมจากวันนี้" else "เลี่ยงจากวันนี้"
                    "birth_day" -> if (positive) "เสริมจากวันเกิด" else "เลี่ยงจากวันเกิด"
                    "progressed_age" -> if (positive) "เสริมจากอายุย่าง" else "เลี่ยงจากอายุย่าง"
                    else -> set.labelTh
                }
                OutfitColorGroupWithBracket(
                    label = displayLabel,
                    colors = set.colors,
                    positive = positive
                )
            }
        }
    }
}

@Composable
private fun OutfitColorGroupWithBracket(
    label: String,
    colors: List<OutfitColor>,
    positive: Boolean
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(3.dp)
    ) {
        Row(
            horizontalArrangement = Arrangement.spacedBy(4.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            colors.forEach { color ->
                Box(
                    modifier = Modifier
                        .size(19.dp)
                        .background(color.hex.toComposeColor(), CircleShape)
                        .border(1.dp, Color.White.copy(alpha = .45f), CircleShape)
                )
            }
        }
        val bracketWidth = if (colors.size >= 3) 65.dp else if (colors.size == 2) 42.dp else 20.dp
        DownwardBracket(
            modifier = Modifier
                .width(bracketWidth)
                .height(7.dp),
            color = if (positive) GoldPrimary.copy(alpha = 0.6f) else Color(0xFFFF8B8B).copy(alpha = 0.6f)
        )
        Text(
            text = label,
            color = if (positive) warmText else Color(0xFFFFB4B4),
            fontSize = 11.sp,
            fontWeight = FontWeight.Medium,
            maxLines = 1
        )
    }
}

@Composable
private fun DownwardBracket(
    modifier: Modifier = Modifier,
    color: Color = GoldPrimary.copy(alpha = 0.6f),
    strokeWidth: androidx.compose.ui.unit.Dp = 1.dp
) {
    Canvas(modifier = modifier) {
        val strokePx = strokeWidth.toPx()
        val w = size.width
        val h = size.height
        val midX = w / 2f
        val lineY = h - 3.5.dp.toPx()
        val leftX = 2.dp.toPx()
        val rightX = w - 2.dp.toPx()

        val path = Path().apply {
            moveTo(leftX, 0f)
            lineTo(leftX, lineY)
            lineTo(midX - 3.5.dp.toPx(), lineY)
            lineTo(midX, h)
            lineTo(midX + 3.5.dp.toPx(), lineY)
            lineTo(rightX, lineY)
            lineTo(rightX, 0f)
        }

        drawPath(
            path = path,
            color = color,
            style = androidx.compose.ui.graphics.drawscope.Stroke(
                width = strokePx,
                cap = androidx.compose.ui.graphics.StrokeCap.Round,
                join = androidx.compose.ui.graphics.StrokeJoin.Round
            )
        )
    }
}

private fun String.toComposeColor(): Color = runCatching {
    Color(android.graphics.Color.parseColor(this))
}.getOrDefault(Color.Gray)

@Composable
private fun MethodCard(result: AuspiciousSearchResponse) {
    CardBlock {
        Text("หลักที่ใช้", color = GoldLight, fontWeight = FontWeight.Bold)
        Text(result.ruleSets.joinToString(", ").ifBlank { "กฎฤกษ์ตามกิจกรรมที่เลือก" }, color = warmText)
        if (result.activity.ruleScopeNote.isNotBlank()) {
            Spacer(Modifier.height(6.dp))
            Text(result.activity.ruleScopeNote, color = mutedText, fontSize = 13.sp)
        }
        if (result.activity.safetyNote.isNotBlank()) {
            Spacer(Modifier.height(6.dp))
            Text("⚠ ${result.activity.safetyNote}", color = Color(0xFFFFD59A), fontSize = 13.sp)
        }
    }
}

private fun thaiDisplayDate(date: LocalDate): String {
    val thaiMonths = listOf("ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.")
    return "${date.dayOfMonth} ${thaiMonths[date.monthValue - 1]} ${date.year + 543}"
}

@Composable
private fun DateField(
    label: String,
    value: String,
    allowClear: Boolean = false,
    preferYearFirst: Boolean = false,
    minimumDate: LocalDate? = null,
    maximumDate: LocalDate? = null,
    placeholder: String = "แตะเพื่อเลือก",
    onSelected: (String) -> Unit
) {
    val parsed = value.takeIf { it.isNotBlank() }?.let { runCatching { LocalDate.parse(it) }.getOrNull() }
    var showPicker by remember { mutableStateOf(false) }
    val minSelectableDate = minimumDate ?: LocalDate.of(1900, 1, 1)
    val maxSelectableDate = maximumDate ?: LocalDate.of(2100, 12, 31)

    Surface(
        onClick = { showPicker = true },
        modifier = Modifier
            .fillMaxWidth()
            .height(72.dp),
        shape = RoundedCornerShape(13.dp),
        color = lvInner,
        border = BorderStroke(1.2.dp, GoldPrimary.copy(alpha = .38f))
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 10.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(34.dp)
                    .background(GoldPrimary.copy(alpha = 0.16f), CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    Icons.Default.CalendarMonth,
                    null,
                    tint = GoldPrimary,
                    modifier = Modifier.size(18.dp)
                )
            }
            Spacer(Modifier.width(8.dp))
            Column(Modifier.weight(1f), horizontalAlignment = Alignment.Start) {
                Text(
                    text = label,
                    color = GoldLight.copy(alpha = 0.85f),
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Spacer(Modifier.height(2.dp))
                Text(
                    text = parsed?.let(::thaiDisplayDate) ?: placeholder,
                    color = if (parsed != null) Color.White else warmText.copy(alpha = 0.82f),
                    fontSize = 13.5.sp,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    softWrap = false,
                    overflow = TextOverflow.Ellipsis
                )
            }
            if (allowClear && value.isNotBlank()) {
                IconButton(onClick = { onSelected("") }, modifier = Modifier.size(30.dp)) {
                    Icon(Icons.Default.Close, "ล้าง", tint = mutedText, modifier = Modifier.size(16.dp))
                }
            }
        }
    }

    if (showPicker) {
        ThaiCalendarPickerSheet(
            title = label,
            initialDate = (parsed ?: LocalDate.now()).coerceIn(minSelectableDate, maxSelectableDate),
            hasInitialSelection = parsed != null,
            startWithYearSelection = preferYearFirst && parsed == null,
            minimumDate = minSelectableDate,
            maximumDate = maxSelectableDate,
            onDismiss = { showPicker = false },
            onConfirm = {
                onSelected(it.toString())
                showPicker = false
            }
        )
    }
}

@Composable
private fun BirthDateField(value: String, onSelected: (String) -> Unit) {
    val today = remember { LocalDate.now() }
    DateField(
        label = "วันเกิด (พ.ศ. · จำเป็นสำหรับทักษาจร)",
        value = value,
        preferYearFirst = true,
        minimumDate = LocalDate.of(1900, 1, 1),
        maximumDate = today,
        placeholder = "แตะเพื่อเลือกวันเกิด",
        onSelected = onSelected
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun TimeField(value: String, onSelected: (String) -> Unit) {
    val parts = value.split(":")
    val hour = parts.getOrNull(0)?.toIntOrNull() ?: 9
    val minute = parts.getOrNull(1)?.toIntOrNull() ?: 0
    var showPicker by remember { mutableStateOf(false) }
    val hasTime = value.isNotBlank()

    val magicTransition = rememberInfiniteTransition(label = "magicTimeField")
    val shimmerOffset by magicTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1200f,
        animationSpec = infiniteRepeatable(
            animation = tween(durationMillis = 3200, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "magicTimeShimmer"
    )

    val magicBorderBrush = Brush.horizontalGradient(
        colors = listOf(
            GoldPrimary,
            GoldLight,
            Color(0xFFFFDF85),
            Color(0xFFA37C10),
            GoldPrimary
        ),
        startX = shimmerOffset,
        endX = shimmerOffset + 600f
    )

    val magicBgBrush = Brush.horizontalGradient(
        colors = listOf(
            Color(0xFF26160E),
            Color(0xFF160C08),
            Color(0xFF26160E)
        )
    )

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(72.dp)
            .background(magicBgBrush, RoundedCornerShape(13.dp))
            .border(
                BorderStroke(if (hasTime) 1.8.dp else 1.2.dp, magicBorderBrush),
                RoundedCornerShape(13.dp)
            )
            .clickable { showPicker = true }
            .padding(horizontal = 14.dp, vertical = 8.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxSize(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(38.dp)
                    .background(GoldPrimary.copy(alpha = 0.16f), CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    Icons.Default.Schedule,
                    contentDescription = null,
                    tint = GoldPrimary,
                    modifier = Modifier.size(22.dp)
                )
            }
            Spacer(Modifier.width(12.dp))
            Column(Modifier.weight(1f), horizontalAlignment = Alignment.Start) {
                Text(
                    text = "✨ เวลาเกิด · กุญแจสำคัญของทักษาจร",
                    color = GoldLight,
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Bold
                )
                Text(
                    text = value.ifBlank { "แตะเพื่อระบุเวลาเกิด" },
                    color = if (hasTime) Color(0xFFFFFFFF) else warmText.copy(alpha = 0.82f),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
            }
            if (value.isNotBlank()) {
                IconButton(onClick = { onSelected("") }, modifier = Modifier.size(36.dp)) {
                    Icon(Icons.Default.Close, "ล้างเวลาเกิด", tint = mutedText)
                }
            }
        }
    }

    if (showPicker) {
        val pickerState = rememberTimePickerState(initialHour = hour, initialMinute = minute, is24Hour = true)
        ModalBottomSheet(
            onDismissRequest = { showPicker = false },
            sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
            containerColor = Color(0xFF160C08)
        ) {
            Column(
                modifier = Modifier.fillMaxWidth().padding(start = 20.dp, end = 20.dp, bottom = 28.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Text("เลือกเวลาเกิด", color = GoldLight, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                TimePicker(state = pickerState)
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                    TextButton(onClick = { showPicker = false }) { Text("ยกเลิก", color = mutedText) }
                    TextButton(onClick = {
                        onSelected("%02d:%02d".format(pickerState.hour, pickerState.minute))
                        showPicker = false
                    }) { Text("ตกลง", color = GoldPrimary, fontWeight = FontWeight.Bold) }
                }
            }
        }
    }
}

private enum class CalendarPickerStep { Year, Month, Day }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun ThaiCalendarPickerSheet(
    title: String,
    initialDate: LocalDate,
    hasInitialSelection: Boolean,
    startWithYearSelection: Boolean,
    minimumDate: LocalDate,
    maximumDate: LocalDate,
    onDismiss: () -> Unit,
    onConfirm: (LocalDate) -> Unit
) {
    var selectedDate by remember { mutableStateOf(initialDate) }
    var hasPickedDay by remember { mutableStateOf(hasInitialSelection) }
    var visibleMonth by remember { mutableStateOf(YearMonth.from(initialDate)) }
    var pickerStep by remember {
        mutableStateOf(if (startWithYearSelection) CalendarPickerStep.Year else CalendarPickerStep.Day)
    }
    val minimumMonth = YearMonth.from(minimumDate)
    val maximumMonth = YearMonth.from(maximumDate)
    val weekdays = listOf("อา", "จ", "อ", "พ", "พฤ", "ศ", "ส")

    fun selectYear(gregorianYear: Int) {
        visibleMonth = YearMonth.of(gregorianYear, visibleMonth.monthValue).coerceIn(minimumMonth, maximumMonth)
        pickerStep = CalendarPickerStep.Month
    }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
        containerColor = emeraldCard
    ) {
        Column(
            modifier = Modifier.fillMaxWidth().padding(start = 18.dp, end = 18.dp, bottom = 28.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            // Header Section
            Column(
                modifier = Modifier.fillMaxWidth(),
                verticalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = title,
                        color = GoldLight,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.weight(1f)
                    )
                    IconButton(
                        onClick = onDismiss,
                        modifier = Modifier.size(32.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Close,
                            contentDescription = "ปิด",
                            tint = mutedText
                        )
                    }
                }
                Text(
                    text = if (hasPickedDay) "วันที่เลือก: ${thaiDisplayDate(selectedDate)}" else "เลือกปี → เดือน → วัน",
                    color = if (hasPickedDay) warmText else GoldPrimary,
                    fontSize = if (hasPickedDay) 22.sp else 16.sp,
                    fontWeight = FontWeight.Bold
                )
            }

            HorizontalDivider(
                color = GoldPrimary.copy(alpha = 0.35f),
                thickness = 1.dp
            )

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                // Step 1: ปี พ.ศ.
                val isYearActive = pickerStep == CalendarPickerStep.Year
                Button(
                    onClick = { pickerStep = CalendarPickerStep.Year },
                    modifier = Modifier.weight(1f).height(58.dp),
                    border = BorderStroke(
                        if (isYearActive) 1.5.dp else 1.dp,
                        if (isYearActive) GoldLight else GoldPrimary.copy(alpha = .35f)
                    ),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (isYearActive) GoldPrimary else lvInner,
                        contentColor = if (isYearActive) Color(0xFF140B07) else warmText
                    ),
                    shape = RoundedCornerShape(12.dp),
                    contentPadding = PaddingValues(horizontal = 6.dp, vertical = 4.dp)
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            "1. ปี พ.ศ.",
                            color = if (isYearActive) Color(0xFF140B07).copy(alpha = 0.85f) else mutedText,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            "${visibleMonth.year + 543}",
                            color = if (isYearActive) Color(0xFF140B07) else warmText,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = 14.sp,
                            maxLines = 1
                        )
                    }
                }

                // Step 2: เดือน
                val isMonthActive = pickerStep == CalendarPickerStep.Month
                Button(
                    onClick = { pickerStep = CalendarPickerStep.Month },
                    modifier = Modifier.weight(1f).height(58.dp),
                    border = BorderStroke(
                        if (isMonthActive) 1.5.dp else 1.dp,
                        if (isMonthActive) GoldLight else GoldPrimary.copy(alpha = .35f)
                    ),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (isMonthActive) GoldPrimary else lvInner,
                        contentColor = if (isMonthActive) Color(0xFF140B07) else warmText
                    ),
                    shape = RoundedCornerShape(12.dp),
                    contentPadding = PaddingValues(horizontal = 6.dp, vertical = 4.dp)
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            "2. เดือน",
                            color = if (isMonthActive) Color(0xFF140B07).copy(alpha = 0.85f) else mutedText,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            thaiMonths[visibleMonth.monthValue - 1],
                            color = if (isMonthActive) Color(0xFF140B07) else warmText,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = 13.sp,
                            maxLines = 1
                        )
                    }
                }

                // Step 3: วัน
                val isDayActive = pickerStep == CalendarPickerStep.Day
                Button(
                    onClick = { pickerStep = CalendarPickerStep.Day },
                    modifier = Modifier.weight(1f).height(58.dp),
                    border = BorderStroke(
                        if (isDayActive) 1.5.dp else 1.dp,
                        if (isDayActive) GoldLight else GoldPrimary.copy(alpha = .35f)
                    ),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (isDayActive) GoldPrimary else lvInner,
                        contentColor = if (isDayActive) Color(0xFF140B07) else warmText
                    ),
                    shape = RoundedCornerShape(12.dp),
                    contentPadding = PaddingValues(horizontal = 6.dp, vertical = 4.dp)
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            "3. วัน",
                            color = if (isDayActive) Color(0xFF140B07).copy(alpha = 0.85f) else mutedText,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            if (hasPickedDay && selectedDate.year == visibleMonth.year && selectedDate.monthValue == visibleMonth.monthValue) "วันที่ ${selectedDate.dayOfMonth}" else "เลือกวัน",
                            color = if (isDayActive) Color(0xFF140B07) else if (hasPickedDay) warmText else GoldLight,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = 13.sp,
                            maxLines = 1
                        )
                    }
                }
            }

            val arrowFraction by animateFloatAsState(
                targetValue = when (pickerStep) {
                    CalendarPickerStep.Year -> 1f / 6f
                    CalendarPickerStep.Month -> 3f / 6f
                    CalendarPickerStep.Day -> 5f / 6f
                },
                animationSpec = spring(
                    dampingRatio = Spring.DampingRatioLowBouncy,
                    stiffness = Spring.StiffnessMediumLow
                ),
                label = "calendarStepArrow"
            )

            Box(
                modifier = Modifier.fillMaxWidth()
            ) {
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                    colors = CardDefaults.cardColors(containerColor = lvCard),
                    border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.35f)),
                    shape = RoundedCornerShape(16.dp)
                ) {
                    Column(
                        modifier = Modifier.padding(14.dp),
                        verticalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        when (pickerStep) {
                        CalendarPickerStep.Year -> {
                            Text("1  เลือกปี พ.ศ.", color = GoldLight, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                            Text("แตะเลือกปี พ.ศ. ด้านล่างเพื่อไปยังขั้นตอนเลือกเดือน", color = mutedText, fontSize = 12.sp)

                            val years = remember(minimumDate.year, maximumDate.year) {
                                (minimumDate.year..maximumDate.year).reversed().toList()
                            }
                            val selectedYearIndex = years.indexOf(visibleMonth.year).coerceAtLeast(0)
                            val yearGridState = rememberLazyGridState(
                                initialFirstVisibleItemIndex = (selectedYearIndex - 3).coerceAtLeast(0)
                            )
                            LazyVerticalGrid(
                                columns = GridCells.Fixed(3),
                                state = yearGridState,
                                modifier = Modifier.fillMaxWidth().height(260.dp),
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                                verticalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                items(years, key = { it }) { year ->
                                    val selected = year == visibleMonth.year
                                    OutlinedButton(
                                        onClick = { selectYear(year) },
                                        border = BorderStroke(1.dp, if (selected) GoldPrimary else GoldPrimary.copy(alpha = .25f)),
                                        colors = ButtonDefaults.outlinedButtonColors(
                                            containerColor = if (selected) GoldPrimary.copy(alpha = .22f) else Color.Transparent,
                                            contentColor = if (selected) GoldLight else warmText
                                        ),
                                        shape = RoundedCornerShape(10.dp),
                                        contentPadding = PaddingValues(vertical = 10.dp)
                                    ) {
                                        Text(
                                            "${year + 543}",
                                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Medium,
                                            fontSize = 14.sp
                                        )
                                    }
                                }
                            }
                            Button(
                                onClick = { pickerStep = CalendarPickerStep.Month },
                                modifier = Modifier.fillMaxWidth().height(44.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary.copy(alpha = 0.2f), contentColor = GoldLight),
                                border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.5f)),
                                shape = RoundedCornerShape(10.dp)
                            ) {
                                Text("ไปเลือกเดือน (พ.ศ. ${visibleMonth.year + 543}) →", fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                            }
                        }

                        CalendarPickerStep.Month -> {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text("2  เลือกเดือน", color = GoldLight, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                                Text("ปี พ.ศ. ${visibleMonth.year + 543}", color = mutedText, fontSize = 13.sp)
                            }
                            Text("แตะเลือกเดือนด้านล่างเพื่อไปยังขั้นตอนเลือกวัน", color = mutedText, fontSize = 12.sp)
                            LazyVerticalGrid(
                                columns = GridCells.Fixed(3),
                                modifier = Modifier.fillMaxWidth().height(220.dp),
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                                verticalArrangement = Arrangement.spacedBy(8.dp),
                                userScrollEnabled = false
                            ) {
                                items(12) { index ->
                                    val month = index + 1
                                    val candidate = YearMonth.of(visibleMonth.year, month)
                                    val enabled = candidate in minimumMonth..maximumMonth
                                    val selected = month == visibleMonth.monthValue
                                    OutlinedButton(
                                        onClick = {
                                            visibleMonth = candidate
                                            pickerStep = CalendarPickerStep.Day
                                        },
                                        enabled = enabled,
                                        border = BorderStroke(1.dp, if (selected) GoldPrimary else GoldPrimary.copy(alpha = .25f)),
                                        colors = ButtonDefaults.outlinedButtonColors(
                                            containerColor = if (selected) GoldPrimary.copy(alpha = .18f) else Color.Transparent,
                                            contentColor = warmText,
                                            disabledContentColor = mutedText.copy(alpha = .3f)
                                        ),
                                        contentPadding = PaddingValues(horizontal = 4.dp)
                                    ) { Text(thaiMonths[index], fontSize = 13.sp, maxLines = 1) }
                                }
                            }

                            Button(
                                onClick = { pickerStep = CalendarPickerStep.Day },
                                modifier = Modifier.fillMaxWidth().height(44.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary.copy(alpha = 0.2f), contentColor = GoldLight),
                                border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.5f)),
                                shape = RoundedCornerShape(10.dp)
                            ) {
                                Text("ไปเลือกวัน (${thaiMonths[visibleMonth.monthValue - 1]} ${visibleMonth.year + 543}) →", fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                            }
                        }

                        CalendarPickerStep.Day -> {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.SpaceBetween
                            ) {
                                IconButton(
                                    onClick = { visibleMonth = visibleMonth.minusMonths(1) },
                                    enabled = visibleMonth > minimumMonth
                                ) { Icon(Icons.Default.ChevronLeft, "เดือนก่อนหน้า", tint = if (visibleMonth > minimumMonth) GoldPrimary else mutedText.copy(alpha = .3f)) }
                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                    Text("3  เลือกวัน", color = GoldLight, fontWeight = FontWeight.Bold)
                                    Text("${thaiMonths[visibleMonth.monthValue - 1]} ${visibleMonth.year + 543}", color = mutedText, fontSize = 12.sp)
                                }
                                IconButton(
                                    onClick = { visibleMonth = visibleMonth.plusMonths(1) },
                                    enabled = visibleMonth < maximumMonth
                                ) { Icon(Icons.Default.ChevronRight, "เดือนถัดไป", tint = if (visibleMonth < maximumMonth) GoldPrimary else mutedText.copy(alpha = .3f)) }
                            }

                            Row(Modifier.fillMaxWidth()) {
                                weekdays.forEach { day ->
                                    Text(
                                        day,
                                        modifier = Modifier.weight(1f),
                                        color = mutedText,
                                        textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                            }

                            val firstOffset = visibleMonth.atDay(1).dayOfWeek.value % 7
                            val daysInMonth = visibleMonth.lengthOfMonth()
                            LazyVerticalGrid(
                                columns = GridCells.Fixed(7),
                                modifier = Modifier.fillMaxWidth().height(285.dp),
                                userScrollEnabled = false
                            ) {
                                items(42) { cell ->
                                    val dayNumber = cell - firstOffset + 1
                                    if (dayNumber in 1..daysInMonth) {
                                        val date = visibleMonth.atDay(dayNumber)
                                        val enabled = date in minimumDate..maximumDate
                                        val selected = hasPickedDay && date == selectedDate
                                        Box(
                                            modifier = Modifier
                                                .aspectRatio(1f)
                                                .padding(2.dp)
                                                .background(if (selected) GoldPrimary else Color.Transparent, CircleShape)
                                                .clickable(enabled = enabled) {
                                                    selectedDate = date
                                                    hasPickedDay = true
                                                },
                                            contentAlignment = Alignment.Center
                                        ) {
                                            Text(
                                                dayNumber.toString(),
                                                color = when {
                                                    selected -> Color(0xFF140B07)
                                                    enabled -> warmText
                                                    else -> mutedText.copy(alpha = .28f)
                                                },
                                                fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal
                                            )
                                        }
                                    } else {
                                        Spacer(Modifier.aspectRatio(1f))
                                    }
                                }
                            }
                        }
                    }
                }
            }

                // Speech Bubble Caret Pointer
                Canvas(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(9.dp)
                ) {
                    val arrowCenterX = size.width * arrowFraction
                    val arrowWidth = 18.dp.toPx()
                    val arrowHeight = 9.dp.toPx()

                    val path = Path().apply {
                        moveTo(arrowCenterX, 0f)
                        lineTo(arrowCenterX + arrowWidth / 2, arrowHeight)
                        lineTo(arrowCenterX - arrowWidth / 2, arrowHeight)
                        close()
                    }

                    drawPath(
                        path = path,
                        color = lvCard
                    )

                    val strokeWidth = 1.2.dp.toPx()
                    drawLine(
                        color = GoldPrimary.copy(alpha = 0.35f),
                        start = Offset(arrowCenterX - arrowWidth / 2, arrowHeight),
                        end = Offset(arrowCenterX, 0f),
                        strokeWidth = strokeWidth
                    )
                    drawLine(
                        color = GoldPrimary.copy(alpha = 0.35f),
                        start = Offset(arrowCenterX, 0f),
                        end = Offset(arrowCenterX + arrowWidth / 2, arrowHeight),
                        strokeWidth = strokeWidth
                    )
                }
            }

            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                TextButton(onClick = onDismiss) { Text("ยกเลิก", color = mutedText) }
                TextButton(onClick = { onConfirm(selectedDate) }, enabled = hasPickedDay) {
                    Text(
                        if (hasPickedDay) "ตกลง" else "เลือกวันก่อน",
                        color = if (hasPickedDay) GoldPrimary else mutedText.copy(alpha = .45f),
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}
