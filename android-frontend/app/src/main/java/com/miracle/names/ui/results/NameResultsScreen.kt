package com.miracle.names.ui.results

import androidx.compose.foundation.background
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.luminance
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.ui.theme.*
import com.miracle.names.viewmodel.NameSearchViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NameResultsScreen(
    results: List<NameResultItem>,
    mode: String,
    viewModel: NameSearchViewModel
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        // Top App Bar
        SmallTopAppBar(
            title = {
                Text(
                    text = "🏆 ผลจัดอันดับ 100 ชื่อมงคล",
                    color = Color.White,
                    fontWeight = FontWeight.Bold,
                    fontSize = 18.sp
                )
            },
            navigationIcon = {
                IconButton(onClick = { viewModel.resetToForm() }) {
                    Icon(
                        imageVector = Icons.Default.ArrowBack,
                        contentDescription = "ย้อนกลับ",
                        tint = Color.White
                    )
                }
            },
            colors = TopAppBarDefaults.smallTopAppBarColors(containerColor = PurpleHeader)
        )

        // Sub-header summary info
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            colors = CardDefaults.cardColors(containerColor = CardBgDark)
        ) {
            Column(
                modifier = Modifier.padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                Text(
                    text = "📌 โหมด: ${if (mode == "combined") "รวมชื่อ+นามสกุล" else "เฉพาะชื่อตัว"}",
                    color = GoldLight,
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Bold
                )
                val displayWish = if (viewModel.selectedWish.value.isNotEmpty()) viewModel.selectedWish.value else viewModel.prototypeName.value
                if (displayWish.isNotEmpty()) {
                    Text(
                        text = "💬 AI Wish: $displayWish",
                        color = Color.White,
                        fontSize = 12.sp
                    )
                }

                val semWeightPercent = (viewModel.semanticWeight.value * 100).toInt()
                val numWeightPercent = 100 - semWeightPercent
                Text(
                    text = "⚖️ สัดส่วนน้ำหนัก: ⭐ เลขศาสตร์ ${numWeightPercent}%   🎯 ความหมาย ${semWeightPercent}%",
                    color = GoldLight.copy(alpha = 0.9f),
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium
                )
            }
        }

        // Curated recommendations first, followed by the complete ranked list.
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
            contentPadding = PaddingValues(bottom = 24.dp)
        ) {
            if (results.isNotEmpty()) {
                item {
                    RecommendationHeader()
                }
            }
            itemsIndexed(results.take(5)) { index, item ->
                NameItemCard(
                    rank = index + 1,
                    item = item,
                    isFeatured = true,
                    onClick = { viewModel.selectNameForBlessingCard(item) }
                )
            }
            if (results.size > 5) {
                item {
                    Text(
                        text = "รายชื่อทั้งหมด",
                        color = GoldLight,
                        fontWeight = FontWeight.Bold,
                        fontSize = 15.sp,
                        modifier = Modifier.padding(top = 12.dp, bottom = 2.dp)
                    )
                }
            }
            itemsIndexed(results.drop(5)) { index, item ->
                NameItemCard(
                    rank = index + 6,
                    item = item,
                    onClick = { viewModel.selectNameForBlessingCard(item) }
                )
            }
        }
    }
}

@Composable
private fun RecommendationHeader() {
    Card(
        colors = CardDefaults.cardColors(containerColor = GoldPrimary.copy(alpha = 0.16f)),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(
                text = "✨ 5 ชื่อที่น่าเลือกสำหรับคุณ",
                color = GoldLight,
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(3.dp))
            Text(
                text = "คัดจากคะแนนสมดุลและเหตุผลที่ตรวจสอบได้",
                color = Color.White,
                fontSize = 12.sp
            )
        }
    }
}

@Composable
fun NameItemCard(
    rank: Int,
    item: NameResultItem,
    isFeatured: Boolean = false,
    onClick: () -> Unit
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val isDarkSurface = surfaceColor.luminance() < 0.5f
    val cardContainerColor = when {
        isFeatured && isDarkSurface -> Color(0xFF33233D)
        isFeatured -> GoldPrimary.copy(alpha = 0.08f)
        else -> surfaceColor
    }
    val primaryTextColor = MaterialTheme.colorScheme.onSurface
    val secondaryTextColor = primaryTextColor.copy(alpha = if (isDarkSurface) 0.72f else 0.68f)
    val reasonTextColor = when {
        isFeatured && isDarkSurface -> GoldLight
        isFeatured -> TextGold
        else -> secondaryTextColor
    }
    val scoreBadgeColor = if (isDarkSurface) PurpleAccent.copy(alpha = 0.34f) else PurpleAccent.copy(alpha = 0.12f)
    val scoreTextColor = if (isDarkSurface) GoldLight else PurpleAccent

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = cardContainerColor
        ),
        border = if (isFeatured) BorderStroke(1.dp, GoldPrimary.copy(alpha = if (isDarkSurface) 0.42f else 0.32f)) else null,
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Rank Badge
            Box(
                modifier = Modifier
                    .size(38.dp)
                    .clip(RoundedCornerShape(19.dp))
                    .background(if (rank <= 3) GoldPrimary else PurpleHeader),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "#$rank",
                    color = if (rank <= 3) Color.Black else Color.White,
                    fontWeight = FontWeight.Bold,
                    fontSize = 13.sp
                )
            }

            Spacer(modifier = Modifier.width(12.dp))

            // Main Content Area
            Column(modifier = Modifier.weight(1f)) {
                // Top Row: Full Name + Star Badge + Score Badge
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        modifier = Modifier.weight(1f, fill = false)
                    ) {
                        Text(
                            text = item.thName,
                            fontWeight = FontWeight.Bold,
                            fontSize = 20.sp,
                            color = primaryTextColor
                        )
                        if (item.isPerfectCombined) {
                            Spacer(modifier = Modifier.width(6.dp))
                            Icon(
                                imageVector = Icons.Default.Star,
                                contentDescription = "มงคลสมบูรณ์แบบ",
                                tint = GoldPrimary,
                                modifier = Modifier.size(18.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.width(8.dp))

                    Surface(
                        color = scoreBadgeColor,
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Text(
                            text = "🏆 Score: ${String.format("%.1f", item.finalScore)}",
                            fontWeight = FontWeight.Bold,
                            fontSize = 12.sp,
                            color = scoreTextColor,
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp)
                        )
                    }
                }

                Spacer(modifier = Modifier.height(4.dp))

                // Meaning Text
                Text(
                    text = item.meaning,
                    fontSize = 13.sp,
                    color = secondaryTextColor,
                    maxLines = 1,
                    overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis
                )

                if (item.rankingReasons.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(5.dp))
                    Text(
                        text = "✓ ${item.rankingReasons.take(if (isFeatured) 3 else 2).joinToString(" • ")}",
                        fontSize = 11.sp,
                        color = reasonTextColor,
                        maxLines = if (isFeatured) 2 else 1,
                        overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis
                    )
                }

                Spacer(modifier = Modifier.height(8.dp))

                // Bottom Row: Badges (PairCircles)
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    if (item.satPairsDisplay.isNotEmpty()) {
                        PairCircles(pairsStr = item.satPairsDisplay, typesStr = item.satPairType)
                    }
                    if (item.shaSum > 0 && item.shaPairsDisplay.isNotEmpty()) {
                        Spacer(modifier = Modifier.width(4.dp))
                        PairCircles(pairsStr = item.shaPairsDisplay, typesStr = item.shaPairType)
                    }
                }
            }

            Spacer(modifier = Modifier.width(4.dp))

            Icon(
                imageVector = Icons.Default.ChevronRight,
                contentDescription = "ดูการ์ดมงคล",
                tint = secondaryTextColor
            )
        }
    }
}


@Composable
fun PairCircles(pairsStr: String, typesStr: String) {
    val pairs = pairsStr.split(Regex("[/,]")).map { it.trim() }.filter { it.isNotEmpty() }
    val types = typesStr.split(Regex("[/,]")).map { it.trim() }.filter { it.isNotEmpty() }

    Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
        pairs.forEachIndexed { index, pair ->
            val type = types.getOrNull(index) ?: ""
            val bgColor = when {
                type.contains("D10") -> Color(0xFF1B5E20) // Dark Green
                type.contains("D8") || type == "D" -> Color(0xFF4CAF50) // Green (default for D)
                type.contains("D5") -> Color(0xFF81C784) // Light Green
                type.contains("R10") -> Color(0xFFE57373) // Light Red
                type.contains("R7") || type == "R" -> Color(0xFFF44336) // Red (default for R)
                type.contains("R5") -> Color(0xFFB71C1C) // Dark Red
                type.contains("B") -> Color(0xFFB71C1C) // Dark Red (Bad)
                else -> Color.Gray
            }
            Box(
                modifier = Modifier
                    .size(24.dp)
                    .clip(androidx.compose.foundation.shape.CircleShape)
                    .background(bgColor),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = pair,
                    color = Color.White,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}
