package com.miracle.names.ui.results

import com.miracle.names.R
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.draw.scale
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.res.painterResource
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.KeyboardArrowUp
import androidx.compose.material.icons.filled.BugReport
import androidx.compose.material.icons.filled.CalendarMonth
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import com.miracle.names.ui.wizard.StepHeaderPattern
import androidx.compose.ui.graphics.luminance
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.ui.components.RadarWeightSelector
import com.miracle.names.ui.theme.*
import com.miracle.names.viewmodel.NameSearchViewModel
import com.miracle.names.viewmodel.UiState
import com.miracle.names.viewmodel.rankingWeightPercents
import com.miracle.names.viewmodel.resolvedRank
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import java.util.Locale
import kotlin.math.roundToInt

@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class)
@Composable
fun NameResultsScreen(
    state: UiState.Success,
    viewModel: NameSearchViewModel,
    onOpenAuspicious: () -> Unit = {},
    onOpenDebugPurchase: (() -> Unit)? = null
) {
    MaterialTheme(colorScheme = DarkColorScheme) {
        val results = state.results
        val snapshot = state.snapshot
        val isUnlocked by viewModel.isUnlocked.collectAsState()
        val liveProductPrice = viewModel.billingManager?.productPrice?.collectAsState()?.value
            ?: "ดูราคาใน Google Play"
        val isReranking by viewModel.isReranking.collectAsState()
        val rerankError by viewModel.rerankError.collectAsState()
        val rankMovements by viewModel.rankMovements.collectAsState()
        val pendingWeights by viewModel.pendingRankingWeights.collectAsState()
        val rerankRevision by viewModel.rerankRevision.collectAsState()
        val listState = rememberLazyListState()
        val coroutineScope = rememberCoroutineScope()
        var topHighlightRevision by remember { mutableIntStateOf(0) }
        var showsTapHint by remember { mutableStateOf(false) }
        var hasShownTapHint by remember { mutableStateOf(false) }
        var resultFilter by remember { mutableStateOf(ResultFilter.ALL) }
        val filteredResults = remember(results, resultFilter) {
            when (resultFilter) {
                ResultFilter.ALL -> results
                ResultFilter.GOOD -> results.filter { !it.hasBadPair }
                ResultFilter.CAUTION -> results.filter { it.hasBadPair }
            }
        }
        val showUpButton by remember {
            derivedStateOf {
                listState.firstVisibleItemIndex > 0 || listState.firstVisibleItemScrollOffset > 240
            }
        }

        LaunchedEffect(results.isNotEmpty()) {
            if (results.isNotEmpty() && !hasShownTapHint) {
                hasShownTapHint = true
                delay(550)
                showsTapHint = true
                delay(3000)
                showsTapHint = false
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
                        text = "🏆 ผลจัดอันดับชื่อดี 4 มิติ",
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
                actions = {
                    IconButton(onClick = onOpenAuspicious) {
                        Icon(
                            imageVector = Icons.Default.CalendarMonth,
                            contentDescription = "ดูฤกษ์ยาม",
                            tint = GoldPrimary
                        )
                    }
                    onOpenDebugPurchase?.let { openDebug ->
                        IconButton(onClick = openDebug) {
                            Icon(
                                imageVector = Icons.Default.BugReport,
                                contentDescription = "เมนู Debug purchase",
                                tint = Color(0xFFFF9800)
                            )
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = PurpleHeader)
            )

            Box(modifier = Modifier.fillMaxSize()) {
                LazyColumn(
                    state = listState,
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(0.dp),
                    contentPadding = PaddingValues(top = 0.dp, bottom = 88.dp)
                ) {
                    item {
                        Card(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 2.dp, vertical = 6.dp),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFF26160E)),
                            shape = RoundedCornerShape(14.dp),
                            border = androidx.compose.foundation.BorderStroke(0.75.dp, GoldPrimary.copy(alpha = 0.30f))
                        ) {
                            Box(modifier = Modifier.fillMaxWidth()) {
                                StepHeaderPattern(
                                    isCompleted = false,
                                    modifier = Modifier.matchParentSize().alpha(0.35f)
                                )
                                Column(
                                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 8.dp),
                                    verticalArrangement = Arrangement.spacedBy(4.dp)
                                ) {
                                    val appliedWeights = snapshot.weights
                                    val requestedWeights = snapshot.requestedWeights
                                    val livePercents = rankingWeightPercents(pendingWeights)
                                    val liveNumPercent = livePercents[0]
                                    val liveSemPercent = livePercents[1]
                                    val liveSoundPercent = livePercents[2]
                                    val liveSimilarityPercent = livePercents[3]
                                    val modeSummary = if (snapshot.isCombined) {
                                        if (snapshot.isFindSurnameMode) {
                                            "ค้นหานามสกุล สำหรับชื่อ: ${snapshot.surname}"
                                        } else {
                                            "ค้นหาชื่อมงคล สำหรับนามสกุล: ${snapshot.surname}"
                                        }
                                    } else {
                                        "ค้นหาเฉพาะชื่อตัว"
                                    }

                                    Column(
                                        modifier = Modifier.fillMaxWidth(),
                                        verticalArrangement = Arrangement.spacedBy(4.dp)
                                    ) {
                                        Text(
                                            text = "📌 $modeSummary · ${formatDayThai(snapshot.dayOfBirth)}",
                                            color = GoldLight,
                                            fontSize = 12.sp,
                                            fontWeight = FontWeight.Bold,
                                            maxLines = 2,
                                            overflow = TextOverflow.Ellipsis
                                        )
                                        Surface(
                                            modifier = Modifier.fillMaxWidth(),
                                            shape = RoundedCornerShape(10.dp),
                                            color = Color.Transparent,
                                            border = BorderStroke(
                                                width = 0.8.dp,
                                                color = GoldPrimary.copy(alpha = 0.45f)
                                            )
                                        ) {
                                            Row(
                                                modifier = Modifier
                                                    .background(
                                                        Brush.horizontalGradient(
                                                            listOf(
                                                                GoldPrimary.copy(alpha = 0.16f),
                                                                Color(0xFF382014).copy(alpha = 0.92f)
                                                            )
                                                        )
                                                    )
                                                    .padding(horizontal = 10.dp, vertical = 8.dp),
                                                verticalAlignment = Alignment.CenterVertically,
                                                horizontalArrangement = Arrangement.spacedBy(9.dp)
                                            ) {
                                             Box(
                                                modifier = Modifier
                                                    .size(28.dp)
                                                    .background(GoldPrimary.copy(alpha = 0.18f), CircleShape)
                                                    .border(1.dp, GoldPrimary.copy(alpha = 0.55f), CircleShape),
                                                contentAlignment = Alignment.Center
                                            ) {
                                                Text(text = "🎯", fontSize = 14.sp)
                                            }
                                            Text(
                                                text = snapshot.wishQuery,
                                                color = Color.White,
                                                fontSize = 11.5.sp,
                                                fontWeight = FontWeight.Medium,
                                                lineHeight = 16.sp,
                                                maxLines = 2,
                                                overflow = TextOverflow.Ellipsis,
                                                modifier = Modifier.weight(1f)
                                            )
                                        }
                                    }

                                    Row(
                                        modifier = Modifier.fillMaxWidth(),
                                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        WeightPercentStack(
                                            numerologyPercent = liveNumPercent,
                                            semanticPercent = liveSemPercent,
                                            soundPercent = liveSoundPercent,
                                            similarityPercent = liveSimilarityPercent,
                                            color = GoldLight.copy(alpha = 0.9f),
                                            modifier = Modifier.width(125.dp)
                                        )

                                        Box(modifier = Modifier.weight(1f)) {
                                            RadarWeightSelector(
                                                numerologyWeight = pendingWeights.numerology,
                                                semanticWeight = pendingWeights.semantic,
                                                soundWeight = pendingWeights.sound,
                                                similarityWeight = pendingWeights.similarity,
                                                onWeightsChanged = { n, sem, snd, sim ->
                                                    viewModel.updateRankingWeightsAndRerank(n, sem, snd, sim)
                                                },
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
                                                    text = "✨ ลากจุดบนกราฟสี่มิติเพื่อปรับ %",
                                                    color = Color.White.copy(alpha = 0.50f),
                                                    fontSize = 9.5.sp
                                                )
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                    item {
                        CompactRankingControls(
                            numerologyWeight = snapshot.weights.numerology,
                            semanticWeight = snapshot.weights.semantic,
                            soundWeight = snapshot.weights.sound,
                            goodCount = results.count { !it.hasBadPair },
                            cautionCount = results.count { it.hasBadPair },
                            selectedFilter = resultFilter,
                            onFilterSelected = { resultFilter = it },
                            isCombined = snapshot.isCombined,
                            isFindSurnameMode = snapshot.isFindSurnameMode,
                            referenceName = snapshot.surname
                        )
                    }
                    stickyHeader {
                        RankingTableHeader(
                            isCombined = snapshot.isCombined,
                            isFindSurnameMode = snapshot.isFindSurnameMode,
                            referenceName = snapshot.surname
                        )
                    }
                    if (!isUnlocked && results.isNotEmpty()) {
                        item {
                            val context = androidx.compose.ui.platform.LocalContext.current
                            val activity = context as? android.app.Activity
                            GoldLifetimeUnlockCard(
                                productPrice = liveProductPrice,
                                onUnlockClick = { viewModel.unlockAllNames(activity) },
                                onRestoreClick = { viewModel.restorePurchases() }
                            )
                        }
                    }
                    if (filteredResults.isEmpty()) {
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
                    itemsIndexed(filteredResults, key = { _, item -> item.nameId }) { index, item ->
                        val originalIndex = results.indexOfFirst { it.nameId == item.nameId }.coerceAtLeast(index)
                        val candidateName = if (!isUnlocked) maskNameCandidate(item.thName) else item.thName
                        val displayName = candidateName
                        NameRankingRow(
                            rank = resolvedRank(item, originalIndex),
                            item = item,
                            displayName = displayName,
                            isCombined = snapshot.isCombined,
                            isFindSurnameMode = snapshot.isFindSurnameMode,
                            rankChange = rankMovements[item.nameId] ?: 0,
                            highlightTop = originalIndex == 0 && topHighlightRevision > 0,
                            isEvenRow = index % 2 == 0,
                            enabled = !isReranking,
                            modifier = Modifier.animateItemPlacement(),
                            onPairClick = { pair -> viewModel.fetchPairDetail(pair) },
                            onClick = { viewModel.selectNameForBlessingCard(item) }
                        )
                    }
                }

                androidx.compose.animation.AnimatedVisibility(
                    visible = showsTapHint,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(start = 16.dp, end = 16.dp, bottom = 18.dp),
                    enter = fadeIn(animationSpec = tween(280)) + slideInVertically(initialOffsetY = { it / 2 }),
                    exit = fadeOut(animationSpec = tween(240)) + slideOutVertically(targetOffsetY = { it / 2 })
                ) {
                    ShimmeringGoldToastBanner(text = "แตะชื่อใดก็ได้เพื่อเปิดการ์ดมงคล")
                }

                androidx.compose.animation.AnimatedVisibility(
                    visible = showUpButton,
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(end = 16.dp, bottom = if (showsTapHint) 82.dp else 18.dp),
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
    similarityPercent: Int,
    color: Color,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(7.dp)
    ) {
        WeightPercentLine(icon = "⭐", label = "เลขศาสตร์", percent = numerologyPercent, color = Color(0xFFFDE047))
        WeightPercentLine(icon = "🎯", label = "ความหมาย", percent = semanticPercent, color = Color(0xFFF2B94B))
        WeightPercentLine(icon = "🔍", label = "เอกลักษณ์", percent = similarityPercent, color = Color(0xFF4DB6AC))
        WeightPercentLine(icon = "🎶", label = "ไพเราะ", percent = soundPercent, color = Color(0xFF7DD3A8))
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
        horizontalArrangement = Arrangement.spacedBy(3.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = icon,
            color = color,
            fontSize = 13.sp,
            maxLines = 1,
            modifier = Modifier.width(18.dp)
        )
        Text(
            text = label,
            color = Color.White.copy(alpha = 0.95f),
            fontSize = 12.sp,
            fontWeight = FontWeight.SemiBold,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f)
        )
        Text(
            text = "$percent%",
            color = color,
            fontSize = 13.sp,
            fontWeight = FontWeight.ExtraBold,
            maxLines = 1,
            modifier = Modifier.width(36.dp)
        )
    }
}

@Composable
fun RankingTableHeader(
    isCombined: Boolean,
    isFindSurnameMode: Boolean = false,
    referenceName: String = ""
) {
    val headerBrush = Brush.horizontalGradient(
        listOf(Color(0xFF1B0F09), Color(0xFF26160E), Color(0xFF1B0F09))
    )
    val labelColor = Color(0xFFFFF1C2)
    val accentLabelColor = GoldLight
    val dividerColor = GoldPrimary.copy(alpha = 0.42f)

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(headerBrush)
    ) {
        if (isCombined && referenceName.isNotBlank()) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFF160C08))
                    .padding(horizontal = 10.dp, vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                Text(
                    text = "🏷️ " + if (isFindSurnameMode) "ค้นหานามสกุลสำหรับชื่อ:" else "ค้นหาชื่อสำหรับนามสกุล:",
                    color = GoldLight,
                    fontSize = 11.sp,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    text = referenceName,
                    color = Color(0xFFFFF1C2),
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold
                )
            }
            HorizontalDivider(color = GoldPrimary.copy(alpha = 0.25f), thickness = 0.5.dp)
        }

        HorizontalDivider(color = GoldPrimary.copy(alpha = 0.34f), thickness = 0.75.dp)
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(42.dp)
                .padding(horizontal = 6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            TableHeaderCell("#", Modifier.width(42.dp), accentLabelColor, Alignment.Center)
            VerticalDivider(color = dividerColor, modifier = Modifier.height(30.dp), thickness = 1.dp)
            TableHeaderCell(
                text = if (isFindSurnameMode) "นามสกุล" else "ชื่อ",
                modifier = Modifier.weight(1f).padding(horizontal = 8.dp),
                color = labelColor,
                alignment = Alignment.CenterStart
            )
            VerticalDivider(color = dividerColor, modifier = Modifier.height(30.dp), thickness = 1.dp)
            TableHeaderCell(
                text = "คู่เลข",
                modifier = Modifier.width(136.dp),
                color = accentLabelColor,
                alignment = Alignment.Center
            )
            VerticalDivider(color = dividerColor, modifier = Modifier.height(30.dp), thickness = 1.dp)
            TableHeaderCell("คะแนน", Modifier.width(52.dp), accentLabelColor, Alignment.CenterEnd)
            Spacer(modifier = Modifier.width(30.dp))
        }
        HorizontalDivider(color = GoldPrimary.copy(alpha = 0.72f), thickness = 1.dp)
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
        Text(
            text = text,
            color = color,
            fontSize = 11.sp,
            fontWeight = FontWeight.Bold,
            letterSpacing = 0.25.sp,
            maxLines = 1
        )
    }
}

internal fun buildRankingDisplayName(
    candidate: String,
    fixedPart: String,
    isCombined: Boolean,
    isFindSurnameMode: Boolean = false
): String {
    val cleanCandidate = candidate.trim()
    val cleanFixedPart = fixedPart.trim()
    if (!isCombined || cleanFixedPart.isBlank()) return cleanCandidate
    return if (isFindSurnameMode) {
        "$cleanFixedPart $cleanCandidate"
    } else {
        "$cleanCandidate $cleanFixedPart"
    }
}

@Composable
fun NameRankingRow(
    rank: Int,
    item: NameResultItem,
    displayName: String = item.thName,
    isCombined: Boolean,
    isFindSurnameMode: Boolean = false,
    rankChange: Int = 0,
    highlightTop: Boolean = false,
    isEvenRow: Boolean = true,
    enabled: Boolean = true,
    modifier: Modifier = Modifier,
    onPairClick: ((String) -> Unit)? = null,
    onClick: () -> Unit
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val isDarkSurface = surfaceColor.luminance() < 0.5f
    val primaryTextColor = MaterialTheme.colorScheme.onSurface
    val secondaryTextColor = primaryTextColor.copy(alpha = if (isDarkSurface) 0.60f else 0.54f)
    val rowContainerColor = when {
        highlightTop -> if (isDarkSurface) Color(0xFF382014) else Color(0xFFFFFBEB)
        isEvenRow -> if (isDarkSurface) Color(0xFF1F120B) else Color(0xFFFFFFFF)
        else -> if (isDarkSurface) Color(0xFF160C08) else Color(0xFFF9FAF9)
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
                text = displayName,
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

            Column(
                modifier = Modifier
                    .width(136.dp)
                    .padding(horizontal = 2.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(1.dp)
            ) {
                if (isCombined && item.satSum > 0) {
                    val candidateStructured = item.candidateSatPairs + item.candidateShaPairs
                    val candidateSatPairs = formatSumToPairs(item.satSum)
                    val candidateShaPairs = formatSumToPairs(item.shaSum)
                    val candidatePairsDisplay = if (candidateStructured.isNotEmpty()) {
                        candidateStructured.joinToString(", ") { it.value }
                    } else (candidateSatPairs + candidateShaPairs).joinToString(", ")
                    val combinedStructured = item.satPairs + item.shaPairs
                    val combinedPairsStr = buildString {
                        if (combinedStructured.isNotEmpty()) append(combinedStructured.joinToString(", ") { it.value })
                        else {
                            if (item.satPairsDisplay.isNotBlank()) append(item.satPairsDisplay)
                            if (item.shaPairsDisplay.isNotBlank()) { if (isNotEmpty()) append(", "); append(item.shaPairsDisplay) }
                        }
                    }
                    val combinedTypesStr = buildString {
                        if (item.satPairType.isNotBlank()) append(item.satPairType)
                        if (item.shaPairType.isNotBlank()) { if (isNotEmpty()) append(", "); append(item.shaPairType) }
                    }
                    // Tier 1 label
                    Text(
                        text = if (isFindSurnameMode) "สกุล" else "ชื่อ",
                        modifier = Modifier.fillMaxWidth(),
                        textAlign = TextAlign.Center,
                        fontSize = 8.sp,
                        fontWeight = FontWeight.Bold,
                        color = GoldLight
                    )
                    // Tier 1 circles — Box guarantees centering
                    Box(modifier = Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                        PairCircles(
                            pairsStr = candidatePairsDisplay,
                            typesStr = "",
                            goodFlags = candidateStructured.takeIf { it.isNotEmpty() }?.map { it.isGood },
                            circleSize = 24.dp,
                            fontSize = 10.sp,
                            strokeWidth = 1.2.dp,
                            spacing = 4.dp,
                            onPairClick = onPairClick
                        )
                    }
                    HorizontalDivider(
                        modifier = Modifier.fillMaxWidth().padding(vertical = 2.dp),
                        thickness = 0.5.dp,
                        color = Color.White.copy(alpha = 0.15f)
                    )
                    // Tier 2 label
                    Text(
                        text = "รวม",
                        modifier = Modifier.fillMaxWidth(),
                        textAlign = TextAlign.Center,
                        fontSize = 8.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White.copy(alpha = 0.85f)
                    )
                    // Tier 2 circles — Box guarantees centering
                    Box(modifier = Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                        PairCircles(
                            pairsStr = combinedPairsStr,
                            typesStr = combinedTypesStr,
                            goodFlags = combinedStructured.takeIf { it.isNotEmpty() }?.map { it.isGood },
                            circleSize = 24.dp,
                            fontSize = 10.sp,
                            strokeWidth = 1.2.dp,
                            spacing = 4.dp,
                            onPairClick = onPairClick
                        )
                    }
                } else {
                    Row(
                        modifier = Modifier.horizontalScroll(rememberScrollState()),
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        PairResultLine(
                            pairs = item.satPairsDisplay,
                            types = item.satPairType,
                            goodFlags = item.satPairs.takeIf { it.isNotEmpty() }?.map { it.isGood },
                            labelColor = secondaryTextColor,
                            onPairClick = onPairClick
                        )
                        PairResultLine(
                            pairs = item.shaPairsDisplay,
                            types = item.shaPairType,
                            goodFlags = item.shaPairs.takeIf { it.isNotEmpty() }?.map { it.isGood },
                            labelColor = secondaryTextColor,
                            onPairClick = onPairClick
                        )
                    }
                }
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
    goodFlags: List<Boolean>? = null,
    labelColor: Color,
    circleSize: androidx.compose.ui.unit.Dp = 36.dp,
    fontSize: androidx.compose.ui.unit.TextUnit = 13.5.sp,
    strokeWidth: androidx.compose.ui.unit.Dp = 1.8.dp,
    spacing: androidx.compose.ui.unit.Dp = 6.dp,
    onPairClick: ((String) -> Unit)? = null
) {
    Row(
        modifier = Modifier.wrapContentWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(4.dp)
    ) {
        if (pairs.isNotBlank()) {
            PairCircles(
                pairsStr = pairs,
                typesStr = types,
                goodFlags = goodFlags,
                circleSize = circleSize,
                fontSize = fontSize,
                strokeWidth = strokeWidth,
                spacing = spacing,
                onPairClick = onPairClick
            )
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
    circleSize: androidx.compose.ui.unit.Dp = 36.dp,
    fontSize: androidx.compose.ui.unit.TextUnit = 13.5.sp,
    strokeWidth: androidx.compose.ui.unit.Dp = 1.8.dp,
    spacing: androidx.compose.ui.unit.Dp = 6.dp,
    onPairClick: ((String) -> Unit)? = null
) {
    val pairs = pairsStr.split(Regex("[/,]")).map { it.trim() }.filter { it.isNotEmpty() }
    val types = typesStr.split(Regex("[/,]")).map { it.trim() }.filter { it.isNotEmpty() }
    val brandGold = GoldPrimary

    Row(modifier = Modifier.wrapContentWidth(), horizontalArrangement = Arrangement.spacedBy(spacing)) {
        pairs.forEachIndexed { index, pair ->
            val type = types.getOrNull(index).orEmpty().uppercase(Locale.ROOT)
            val explicitlyGood = goodFlags?.getOrNull(index)
            val isBad = explicitlyGood == false || type.contains("R") || type.contains("B")
            val bgColor = if (isBad) StatusCaution else StatusPositive
            val statusLabel = if (isBad) "ควรพิจารณา" else "ส่งเสริม"

            Box(
                modifier = Modifier
                    .size(circleSize)
                    .clip(CircleShape)
                    .background(bgColor)
                    .border(BorderStroke(strokeWidth, brandGold), shape = CircleShape)
                    .semantics { contentDescription = "คู่เลข $pair: $statusLabel" }
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
                    fontSize = fontSize,
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

private fun maskNameCandidate(name: String): String {
    if (name.isEmpty()) return name
    val visibleCount = maxOf(2, name.length / 3)
    return name.take(visibleCount) + "***"
}

@Composable
private fun GoldLifetimeUnlockCard(
    productPrice: String,
    onUnlockClick: () -> Unit,
    onRestoreClick: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 10.dp, vertical = 8.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color(0xFF1F120B)),
        border = androidx.compose.foundation.BorderStroke(
            1.5.dp,
            androidx.compose.ui.graphics.Brush.horizontalGradient(
                listOf(
                    GoldPrimary,
                    Color(0xFFFDE047),
                    GoldPrimary
                )
            )
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 6.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(14.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Image(
                painter = painterResource(id = R.drawable.app_logo),
                contentDescription = "โลโก้ชื่อดี",
                modifier = Modifier
                    .size(34.dp)
                    .clip(CircleShape)
                    .border(1.dp, GoldPrimary.copy(alpha = 0.65f), CircleShape)
            )
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "100 อันดับชื่อที่ดีที่สุด",
                    color = GoldLight,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold
                )
                Text(
                    text = "เปิดชื่อทั้งหมดแบบถาวร",
                    color = Color.White.copy(alpha = 0.88f),
                    fontSize = 11.sp,
                    lineHeight = 15.sp
                )
                Surface(
                    shape = RoundedCornerShape(999.dp),
                    color = GoldPrimary.copy(alpha = 0.16f),
                    border = BorderStroke(0.8.dp, GoldPrimary.copy(alpha = 0.42f)),
                    modifier = Modifier.padding(top = 5.dp)
                ) {
                    Text(
                        text = productPrice,
                        color = GoldLight,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                    )
                }
            }
            Column(horizontalAlignment = Alignment.End) {
                Button(
                    onClick = onUnlockClick,
                    colors = ButtonDefaults.buttonColors(
                        containerColor = GoldPrimary,
                        contentColor = Color(0xFF1F120B)
                    ),
                    shape = RoundedCornerShape(12.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "เปิดชื่อทั้งหมด",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                if (onRestoreClick != null) {
                    Spacer(modifier = Modifier.height(9.dp))
                    TextButton(
                        onClick = onRestoreClick,
                        modifier = Modifier
                            .border(
                                width = 0.8.dp,
                                color = GoldPrimary.copy(alpha = 0.32f),
                                shape = RoundedCornerShape(999.dp)
                            ),
                        contentPadding = PaddingValues(horizontal = 10.dp, vertical = 3.dp)
                    ) {
                        Text(
                            text = "คืนสิทธิ์การซื้อ",
                            color = GoldLight.copy(alpha = 0.75f),
                            fontSize = 10.sp
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun ShimmeringGoldToastBanner(
    text: String,
    modifier: Modifier = Modifier
) {
    val shimmerAnim = remember { Animatable(-1f) }
    val infiniteTransition = rememberInfiniteTransition(label = "toast_glow")
    val glowAlpha by infiniteTransition.animateFloat(
        initialValue = 0.40f,
        targetValue = 0.90f,
        animationSpec = infiniteRepeatable(
            animation = tween(1100, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "glow_pulse"
    )
    val sparkleScale by infiniteTransition.animateFloat(
        initialValue = 0.92f,
        targetValue = 1.25f,
        animationSpec = infiniteRepeatable(
            animation = tween(800, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "sparkle_scale"
    )

    LaunchedEffect(Unit) {
        shimmerAnim.animateTo(
            targetValue = 2f,
            animationSpec = tween(durationMillis = 1200, easing = LinearEasing)
        )
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .shadow(
                elevation = (12 * glowAlpha).dp,
                shape = RoundedCornerShape(16.dp),
                ambientColor = Color(0xFFFFD700),
                spotColor = Color(0xFFFFB300)
            )
            .clip(RoundedCornerShape(16.dp))
            .background(
                brush = Brush.linearGradient(
                    colors = listOf(
                        Color(0xFFFFE059),
                        Color(0xFFFA9E1F),
                        Color(0xFFFFD64D)
                    )
                )
            )
            .border(
                border = BorderStroke(
                    width = 1.8.dp,
                    brush = Brush.linearGradient(
                        colors = listOf(
                            Color.White,
                            Color(0xFFFFF7B2),
                            GoldPrimary,
                            Color.White
                        )
                    )
                ),
                shape = RoundedCornerShape(16.dp)
            )
            .drawWithContent {
                drawContent()
                val width = size.width
                val xOffset = shimmerAnim.value * width
                drawRect(
                    brush = Brush.linearGradient(
                        colors = listOf(
                            Color.Transparent,
                            Color.White.copy(alpha = 0.12f),
                            Color.White.copy(alpha = 0.80f),
                            Color.White.copy(alpha = 0.12f),
                            Color.Transparent
                        ),
                        start = Offset(xOffset - 120f, 0f),
                        end = Offset(xOffset + 120f, size.height)
                    )
                )
            }
            .padding(horizontal = 18.dp, vertical = 16.dp),
        contentAlignment = Alignment.Center
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = "✨",
                fontSize = 16.sp,
                modifier = Modifier.scale(sparkleScale)
            )
            Text(
                text = text,
                color = Color(0xFF1F120B),
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold
            )
        }
    }
}

private enum class ResultFilter { ALL, GOOD, CAUTION }

@Composable
private fun CompactRankingControls(
    numerologyWeight: Float,
    semanticWeight: Float,
    soundWeight: Float,
    goodCount: Int,
    cautionCount: Int,
    selectedFilter: ResultFilter,
    onFilterSelected: (ResultFilter) -> Unit,
    isCombined: Boolean = false,
    isFindSurnameMode: Boolean = false,
    referenceName: String = "",
    modifier: Modifier = Modifier
) {
    val formulaLabel = remember(numerologyWeight, semanticWeight, soundWeight) {
        val maxVal = maxOf(numerologyWeight, semanticWeight, soundWeight)
        when (maxVal) {
            numerologyWeight -> "เน้นเลขศาสตร์"
            semanticWeight -> "เน้นความหมาย"
            else -> "เน้นความไพเราะ"
        }
    }

    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(GreenHeader.copy(alpha = 0.48f))
            .horizontalScroll(rememberScrollState())
            .padding(horizontal = 10.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(7.dp)
    ) {
        Text(
            text = "สูตร: $formulaLabel",
            color = GoldLight,
            fontSize = 11.sp,
            fontWeight = FontWeight.Bold
        )

        CompactFilterChip(
            title = "ทั้งหมด ${goodCount + cautionCount}",
            isSelected = selectedFilter == ResultFilter.ALL,
            onClick = { onFilterSelected(ResultFilter.ALL) }
        )

        CompactFilterChip(
            title = "คู่ดี $goodCount",
            isSelected = selectedFilter == ResultFilter.GOOD,
            onClick = { onFilterSelected(ResultFilter.GOOD) }
        )

        if (cautionCount > 0) {
            CompactFilterChip(
                title = "🔴 ระวัง $cautionCount",
                isSelected = selectedFilter == ResultFilter.CAUTION,
                onClick = { onFilterSelected(ResultFilter.CAUTION) }
            )
        }
    }
}

@Composable
private fun CompactFilterChip(
    title: String,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    Surface(
        onClick = onClick,
        shape = CircleShape,
        color = if (isSelected) GoldPrimary.copy(alpha = 0.15f) else Color.Transparent,
        border = BorderStroke(
            width = 1.dp,
            color = if (isSelected) GoldPrimary.copy(alpha = 0.65f) else Color.White.copy(alpha = 0.16f)
        )
    ) {
        Text(
            text = title,
            color = if (isSelected) GoldLight else Color.White.copy(alpha = 0.65f),
            fontSize = 11.sp,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 11.dp, vertical = 5.dp)
        )
    }
}

internal fun formatSumToPairs(total: Int): List<String> {
    if (total <= 0) return emptyList()
    val s = total.toString()
    if (s.length == 1) return listOf(String.format(Locale.US, "%02d", total))
    if (s.length == 2) return listOf(s)
    val result = mutableListOf<String>()
    for (i in 0 until s.length - 1) {
        result.add(s.substring(i, i + 2))
    }
    return result
}
