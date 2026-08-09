package com.miracle.names.ui.results

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.KeyboardArrowUp
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.luminance
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.ui.components.TriangularWeightSelector
import com.miracle.names.ui.theme.*
import com.miracle.names.viewmodel.NameSearchViewModel
import com.miracle.names.viewmodel.UiState
import com.miracle.names.viewmodel.resolvedRank
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import java.util.Locale
import kotlin.math.roundToInt

@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class)
@Composable
fun NameResultsScreen(
    state: UiState.Success,
    viewModel: NameSearchViewModel
) {
    MaterialTheme(colorScheme = DarkColorScheme) {
        val results = state.results
        val snapshot = state.snapshot
        val isReranking by viewModel.isReranking.collectAsState()
        val rerankError by viewModel.rerankError.collectAsState()
        val rankMovements by viewModel.rankMovements.collectAsState()
        val pendingWeights by viewModel.pendingRankingWeights.collectAsState()
        val rerankRevision by viewModel.rerankRevision.collectAsState()
        val listState = rememberLazyListState()
        val coroutineScope = rememberCoroutineScope()
        var topHighlightRevision by remember { mutableIntStateOf(0) }
        val showUpButton by remember {
            derivedStateOf {
                listState.firstVisibleItemIndex > 0 || listState.firstVisibleItemScrollOffset > 240
            }
        }

        LaunchedEffect(rerankRevision) {
            if (rerankRevision > 0 && results.isNotEmpty()) {
                val revision = rerankRevision
                topHighlightRevision = revision
                listState.animateScrollToItem(0)
                delay(1800)
                if (topHighlightRevision == revision) {
                    topHighlightRevision = 0
                }
            }
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background)
        ) {
            TopAppBar(
                title = {
                    Text(
                        text = "🏆 ผลจัดอันดับชื่อดี 100%",
                        color = Color.White,
                        fontWeight = FontWeight.Bold,
                        fontSize = 18.sp,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis
                    )
                },
                navigationIcon = {
                    IconButton(onClick = viewModel::resetToForm) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "ย้อนกลับ",
                            tint = Color.White
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = PurpleHeader)
            )

            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = CardBgDark),
                shape = RoundedCornerShape(0.dp)
            ) {
                Column(
                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 8.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    val appliedWeights = snapshot.weights
                    val requestedWeights = snapshot.requestedWeights
                    val liveNumPercent = (pendingWeights.numerology * 100).roundToInt()
                    val liveSemPercent = (pendingWeights.semantic * 100).roundToInt()
                    val liveSoundPercent = (100 - liveNumPercent - liveSemPercent).coerceAtLeast(0)
                    val modeSummary = if (snapshot.isCombined) {
                        if (snapshot.isFindSurnameMode) {
                            "${snapshot.surname} + [หานามสกุล]"
                        } else {
                            "[หาชื่อ] + ${snapshot.surname}"
                        }
                    } else {
                        "เฉพาะชื่อตัว"
                    }

                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        Text(
                            text = "📌 โหมด: $modeSummary · ${formatDayThai(snapshot.dayOfBirth)}",
                            color = GoldLight,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Bold,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis
                        )
                        Text(
                            text = "🎯 ความหมาย: ${snapshot.wishQuery}",
                            color = Color.White,
                            fontSize = 10.sp,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis
                        )

                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            WeightPercentStack(
                                numerologyPercent = liveNumPercent,
                                semanticPercent = liveSemPercent,
                                soundPercent = liveSoundPercent,
                                color = GoldLight.copy(alpha = 0.9f),
                                modifier = Modifier.width(132.dp)
                            )

                            Box(modifier = Modifier.weight(1f)) {
                                TriangularWeightSelector(
                                    numerologyWeight = pendingWeights.numerology,
                                    semanticWeight = pendingWeights.semantic,
                                    soundWeight = pendingWeights.sound,
                                    onWeightsChanged = viewModel::updateRankingWeightsAndRerank,
                                    compact = true,
                                    labelColor = Color.White.copy(alpha = 0.78f),
                                    showLabels = false,
                                    showCornerIcons = true
                                )
                            }
                        }
                    }

                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(20.dp),
                        contentAlignment = Alignment.CenterStart
                    ) {
                        AnimatedContent(
                            targetState = Triple(isReranking, rerankError, !requestedWeights.approximatelyEquals(appliedWeights)),
                            transitionSpec = { fadeIn(animationSpec = tween(180)) togetherWith fadeOut(animationSpec = tween(180)) },
                            label = "rerank_status_slot"
                        ) { (reranking, error, hasPending) ->
                            when {
                                reranking -> {
                                    Row(
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(6.dp)
                                    ) {
                                        CircularProgressIndicator(
                                            modifier = Modifier.size(11.dp),
                                            strokeWidth = 2.dp,
                                            color = GoldPrimary
                                        )
                                        Text(
                                            text = "🔮 กำลังคำนวณและจัดอันดับใหม่...",
                                            color = GoldLight,
                                            fontSize = 9.5.sp,
                                            fontWeight = FontWeight.Medium
                                        )
                                    }
                                }
                                error != null -> {
                                    Text(
                                        text = "⚠️ $error · แสดงผลจากสูตรเดิม",
                                        color = Color(0xFFFFB4AB),
                                        fontSize = 9.5.sp
                                    )
                                }
                                hasPending -> {
                                    Text(
                                        text = "✨ สด: กำลังประมวลผลสูตรที่เลือก...",
                                        color = GoldLight.copy(alpha = 0.85f),
                                        fontSize = 9.5.sp
                                    )
                                }
                                else -> {
                                    Text(
                                        text = "✨ ลากจุดบนกราฟสามเหลี่ยมเพื่อปรับ %",
                                        color = Color.White.copy(alpha = 0.50f),
                                        fontSize = 9.5.sp
                                    )
                                }
                            }
                        }
                    }
                }
            }

            Box(modifier = Modifier.fillMaxSize()) {
                LazyColumn(
                    state = listState,
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(0.dp),
                    contentPadding = PaddingValues(top = 0.dp, bottom = 88.dp)
                ) {
                    stickyHeader {
                        RankingTableHeader(
                            isCombined = snapshot.isCombined,
                            isFindSurnameMode = snapshot.isFindSurnameMode
                        )
                    }
                    if (results.isEmpty()) {
                        item {
                            Card(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(24.dp),
                                colors = CardDefaults.cardColors(containerColor = CardBgDark),
                                shape = RoundedCornerShape(12.dp)
                            ) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(24.dp),
                                    horizontalAlignment = Alignment.CenterHorizontally,
                                    verticalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    Text(
                                        text = "ไม่พบรายชื่อที่ผ่านเกณฑ์ตามเงื่อนไขที่กำหนด",
                                        color = Color.White.copy(alpha = 0.88f),
                                        fontSize = 15.sp,
                                        fontWeight = FontWeight.Medium,
                                        maxLines = 2
                                    )
                                }
                            }
                        }
                    }
                    itemsIndexed(results, key = { _, item -> item.nameId }) { index, item ->
                        NameRankingRow(
                            rank = resolvedRank(item, index),
                            item = item,
                            isCombined = snapshot.isCombined,
                            rankChange = rankMovements[item.nameId] ?: 0,
                            highlightTop = index == 0 && topHighlightRevision > 0,
                            enabled = !isReranking,
                            modifier = Modifier.animateItemPlacement(),
                            onPairClick = { pair -> viewModel.fetchPairDetail(pair) },
                            onClick = { viewModel.selectNameForBlessingCard(item) }
                        )
                    }
                }

                androidx.compose.animation.AnimatedVisibility(
                    visible = showUpButton,
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(end = 16.dp, bottom = 18.dp),
                    enter = fadeIn(animationSpec = tween(180)) + scaleIn(animationSpec = tween(180)),
                    exit = fadeOut(animationSpec = tween(140)) + scaleOut(animationSpec = tween(140))
                ) {
                    FloatingActionButton(
                        onClick = {
                            coroutineScope.launch {
                                listState.animateScrollToItem(0)
                            }
                        },
                        containerColor = GoldPrimary,
                        contentColor = Color.Black
                    ) {
                        Icon(
                            imageVector = Icons.Default.KeyboardArrowUp,
                            contentDescription = "ขึ้นบนสุด",
                            modifier = Modifier.size(28.dp)
                        )
                    }
                }
            }
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
private fun WeightPercentStack(
    numerologyPercent: Int,
    semanticPercent: Int,
    soundPercent: Int,
    color: Color,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(2.dp)
    ) {
        WeightPercentLine(icon = "⭐", label = "เลขศาสตร์", percent = numerologyPercent, color = Color(0xFFFDE047))
        WeightPercentLine(icon = "🎯", label = "ความหมาย", percent = semanticPercent, color = Color(0xFFF472B6))
        WeightPercentLine(icon = "🎶", label = "ไพเราะ", percent = soundPercent, color = Color(0xFF38BDF8))
    }
}

@Composable
private fun WeightPercentLine(
    icon: String,
    label: String,
    percent: Int,
    color: Color
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(5.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = icon,
            color = color,
            fontSize = 10.sp,
            maxLines = 1,
            modifier = Modifier.width(16.dp)
        )
        Text(
            text = label,
            color = color,
            fontSize = 10.sp,
            fontWeight = FontWeight.Medium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f)
        )
        Text(
            text = "$percent%",
            color = color,
            fontSize = 10.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            modifier = Modifier.width(32.dp)
        )
    }
}

@Composable
fun RankingTableHeader(
    isCombined: Boolean,
    isFindSurnameMode: Boolean = false
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val isDarkSurface = surfaceColor.luminance() < 0.5f
    val headerColor = if (isDarkSurface) Color(0xFF21142B) else Color(0xFFF1ECF5)
    val labelColor = MaterialTheme.colorScheme.onSurface.copy(alpha = if (isDarkSurface) 0.72f else 0.62f)
    val dividerColor = MaterialTheme.colorScheme.outline.copy(alpha = if (isDarkSurface) 0.16f else 0.22f)

    Column(modifier = Modifier.fillMaxWidth().background(headerColor)) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(34.dp)
                .padding(horizontal = 6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            TableHeaderCell("#", Modifier.width(42.dp), labelColor, Alignment.Center)
            VerticalDivider(color = dividerColor, modifier = Modifier.height(22.dp), thickness = 0.5.dp)
            TableHeaderCell(
                text = if (isFindSurnameMode) "นามสกุล" else "ชื่อ",
                modifier = Modifier.weight(1f).padding(horizontal = 8.dp),
                color = labelColor,
                alignment = Alignment.CenterStart
            )
            VerticalDivider(color = dividerColor, modifier = Modifier.height(22.dp), thickness = 0.5.dp)
            TableHeaderCell(
                text = if (isCombined) "คู่เลขรวม" else "คู่เลข",
                modifier = Modifier.width(128.dp),
                color = labelColor,
                alignment = Alignment.Center
            )
            VerticalDivider(color = dividerColor, modifier = Modifier.height(22.dp), thickness = 0.5.dp)
            TableHeaderCell("คะแนน", Modifier.width(52.dp), labelColor, Alignment.CenterEnd)
            Spacer(modifier = Modifier.width(30.dp))
        }
        HorizontalDivider(color = dividerColor, thickness = 0.5.dp)
    }
}

@Composable
private fun TableHeaderCell(
    text: String,
    modifier: Modifier,
    color: Color,
    alignment: Alignment
) {
    Box(modifier = modifier, contentAlignment = alignment) {
        Text(text = text, color = color, fontSize = 10.sp, fontWeight = FontWeight.Bold, maxLines = 1)
    }
}

@Composable
fun NameRankingRow(
    rank: Int,
    item: NameResultItem,
    isCombined: Boolean,
    rankChange: Int = 0,
    highlightTop: Boolean = false,
    enabled: Boolean = true,
    modifier: Modifier = Modifier,
    onPairClick: ((String) -> Unit)? = null,
    onClick: () -> Unit
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val isDarkSurface = surfaceColor.luminance() < 0.5f
    val primaryTextColor = MaterialTheme.colorScheme.onSurface
    val secondaryTextColor = primaryTextColor.copy(alpha = if (isDarkSurface) 0.60f else 0.54f)
    val rowContainerColor = if (highlightTop) {
        if (isDarkSurface) Color(0xFF342441) else Color(0xFFFFFBEB)
    } else {
        surfaceColor
    }
    val scoreTextColor = if (isDarkSurface) GoldLight else PurpleAccent
    val animatedContainerColor by animateColorAsState(
        targetValue = rowContainerColor,
        animationSpec = tween(durationMillis = 520),
        label = "ranking-row-container"
    )

    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(animatedContainerColor)
            .clickable(enabled = enabled, onClick = onClick)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .heightIn(min = 66.dp)
                .padding(horizontal = 6.dp, vertical = 5.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(modifier = Modifier.width(42.dp).fillMaxHeight(), contentAlignment = Alignment.Center) {
                Text(
                    text = "#$rank",
                    color = if (rank <= 3) GoldPrimary else secondaryTextColor,
                    fontWeight = FontWeight.Bold,
                    fontSize = 14.sp
                )
            }

            VerticalDivider(
                color = MaterialTheme.colorScheme.outline.copy(alpha = if (isDarkSurface) 0.14f else 0.20f),
                modifier = Modifier.height(42.dp),
                thickness = 0.5.dp
            )

            Text(
                text = item.thName,
                fontWeight = FontWeight.Bold,
                fontSize = 17.sp,
                color = primaryTextColor,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f).padding(horizontal = 8.dp)
            )

            VerticalDivider(
                color = MaterialTheme.colorScheme.outline.copy(alpha = if (isDarkSurface) 0.14f else 0.20f),
                modifier = Modifier.height(42.dp),
                thickness = 0.5.dp
            )

            Row(
                modifier = Modifier
                    .width(112.dp)
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = 5.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                PairResultLine(
                    pairs = item.satPairsDisplay,
                    types = item.satPairType,
                    labelColor = secondaryTextColor,
                    onPairClick = onPairClick
                )
                PairResultLine(
                    pairs = item.shaPairsDisplay,
                    types = item.shaPairType,
                    labelColor = secondaryTextColor,
                    onPairClick = onPairClick
                )
            }

            VerticalDivider(
                color = MaterialTheme.colorScheme.outline.copy(alpha = if (isDarkSurface) 0.14f else 0.20f),
                modifier = Modifier.height(42.dp),
                thickness = 0.5.dp
            )

            Column(
                modifier = Modifier.width(52.dp),
                horizontalAlignment = Alignment.End,
                verticalArrangement = Arrangement.Center
            ) {
                Text(
                    text = String.format(Locale.US, "%.2f", item.finalScore),
                    color = scoreTextColor,
                    fontWeight = FontWeight.Bold,
                    fontSize = 12.sp
                )
                if (rankChange != 0) {
                    Text(
                        text = if (rankChange > 0) "↑$rankChange" else "↓${-rankChange}",
                        color = if (rankChange > 0) Color(0xFF2E7D32) else Color(0xFFC62828),
                        fontSize = 9.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }

            Box(modifier = Modifier.width(30.dp), contentAlignment = Alignment.Center) {
                Icon(
                    imageVector = Icons.Default.ChevronRight,
                    contentDescription = "ดูรายละเอียดคะแนนและการ์ดมงคล",
                    tint = secondaryTextColor,
                    modifier = Modifier.size(18.dp)
                )
            }
        }
        HorizontalDivider(
            color = MaterialTheme.colorScheme.outline.copy(alpha = if (isDarkSurface) 0.18f else 0.28f),
            thickness = 0.5.dp
        )
    }
}

@Composable
private fun PairResultLine(
    pairs: String,
    types: String,
    labelColor: Color,
    onPairClick: ((String) -> Unit)? = null
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(4.dp)
    ) {
        if (pairs.isNotBlank()) {
            PairCircles(pairsStr = pairs, typesStr = types, onPairClick = onPairClick)
        } else {
            Text(text = "ไม่มีข้อมูล", color = labelColor, fontSize = 8.sp, maxLines = 1)
        }
    }
}

@Composable
fun PairCircles(
    pairsStr: String,
    typesStr: String,
    goodFlags: List<Boolean>? = null,
    onPairClick: ((String) -> Unit)? = null
) {
    val pairs = pairsStr.split(Regex("[/,]")).map { it.trim() }.filter { it.isNotEmpty() }
    val types = typesStr.split(Regex("[/,]")).map { it.trim() }.filter { it.isNotEmpty() }

    Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
        pairs.forEachIndexed { index, pair ->
            val type = types.getOrNull(index).orEmpty().uppercase(Locale.ROOT)
            val explicitlyGood = goodFlags?.getOrNull(index)
            val bgColor = when {
                explicitlyGood == false -> Color(0xFFB71C1C)
                type.contains("D10") -> Color(0xFF1B5E20)
                type.contains("D8") || type == "D" -> Color(0xFF4CAF50)
                type.contains("D5") -> Color(0xFF81C784)
                type.contains("R10") -> Color(0xFFE57373)
                type.contains("R7") || type == "R" -> Color(0xFFF44336)
                type.contains("R5") -> Color(0xFFB71C1C)
                type.contains("B") -> Color(0xFFB71C1C)
                else -> Color.Gray
            }
            Box(
                modifier = Modifier
                    .size(24.dp)
                    .clip(androidx.compose.foundation.shape.CircleShape)
                    .background(bgColor)
                    .then(
                        if (onPairClick != null) {
                            Modifier.clickable { onPairClick(pair) }
                        } else Modifier
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = pair,
                    color = Color.White,
                    fontSize = 12.sp,
                    fontFamily = FontFamily.Default,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}

private fun formatDayThai(day: String): String {
    return when (day.lowercase(Locale.ROOT)) {
        "sunday" -> "วันอาทิตย์"
        "monday" -> "วันจันทร์"
        "tuesday" -> "วันอังคาร"
        "wednesday1", "wednesday" -> "วันพุธ (กลางวัน)"
        "wednesday2", "wednesday_night" -> "วันพุธ (กลางคืน)"
        "thursday" -> "วันพฤหัสบดี"
        "friday" -> "วันศุกร์"
        "saturday" -> "วันเสาร์"
        else -> day.ifBlank { "ไม่ระบุ" }
    }
}

private fun scopeLabel(scope: String): String {
    return when (scope.lowercase(Locale.ROOT)) {
        "first_name", "name" -> "ชื่อตัว"
        "full_name", "combined" -> "ชื่อ + นามสกุล"
        "both" -> "ทั้งชื่อตัวและชื่อเต็ม"
        "not_applied" -> "ไม่ได้ใช้"
        else -> "ชื่อตัว"
    }
}

private fun semanticSourceLabel(source: String): String {
    return when (source.lowercase(Locale.ROOT)) {
        "embedding", "openai_embedding", "vector" -> "semantic vector"
        "owner_confirmed_legacy_embedding" -> "vector ที่เจ้าของข้อมูลยืนยัน"
        "seed_vector" -> "vector สำรอง"
        "text", "text_fallback", "keyword" -> "ข้อความ/คำสำคัญ"
        "trigram" -> "ความคล้ายชื่อ (trigram)"
        "not_requested" -> "ไม่ได้ร้องขอ"
        "embedding_missing", "unavailable" -> "ไม่พร้อมใช้"
        "none", "" -> "ไม่ระบุแหล่ง"
        else -> source
    }
}
