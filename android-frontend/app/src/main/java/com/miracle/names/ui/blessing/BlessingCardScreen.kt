package com.miracle.names.ui.blessing

import android.widget.Toast
import androidx.compose.animation.core.*
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Share
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.draw.scale
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asAndroidBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import coil.request.CachePolicy
import coil.request.ImageRequest
import com.miracle.names.R
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.ui.theme.*
import com.miracle.names.utils.ImageUtils
import com.miracle.names.viewmodel.AppliedSearchSnapshot
import com.miracle.names.viewmodel.NameSearchViewModel
import dev.shreyaspatil.capturable.capturable
import dev.shreyaspatil.capturable.controller.rememberCaptureController
import kotlinx.coroutines.launch
import java.util.Locale
import androidx.compose.runtime.ExperimentalComposeApi
import kotlin.math.roundToInt

@OptIn(ExperimentalMaterial3Api::class, ExperimentalComposeUiApi::class, ExperimentalComposeApi::class)
@Composable
fun BlessingCardScreen(
    item: NameResultItem? = null,
    snapshot: AppliedSearchSnapshot,
    viewModel: NameSearchViewModel,
    onBackClick: () -> Unit = { viewModel.selectNameForBlessingCard(null) }
) {
    val selectedItem = item ?: viewModel.selectedItem.collectAsState().value
    val dayOfBirth = snapshot.dayOfBirth
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val captureController = rememberCaptureController()
    var useLocalBuddhaImage by remember(dayOfBirth) { mutableStateOf(false) }
    var buddhaImageUnavailable by remember(dayOfBirth) { mutableStateOf(false) }
    val localBuddhaImageResId = remember(dayOfBirth) { localBuddhaPngResourceFor(dayOfBirth) }
    val buddhaImageData = if (useLocalBuddhaImage) {
        localBuddhaImageResId
    } else {
        "http://10.0.2.2:8080/buddhas/${buddhaImageFileNameFor(dayOfBirth)}?v=${System.currentTimeMillis()}"
    }
    val buddhaImageRequest = remember(dayOfBirth, useLocalBuddhaImage) {
        ImageRequest.Builder(context)
            .data(buddhaImageData)
            .memoryCachePolicy(CachePolicy.DISABLED)
            .diskCachePolicy(CachePolicy.DISABLED)
            .build()
    }

    val frogLeftImageRequest = remember {
        ImageRequest.Builder(context)
            .data("http://10.0.2.2:8080/images/frog_left.png?v=${System.currentTimeMillis()}")
            .error(R.drawable.frog_left)
            .placeholder(R.drawable.frog_left)
            .memoryCachePolicy(CachePolicy.DISABLED)
            .diskCachePolicy(CachePolicy.DISABLED)
            .build()
    }

    val goldRightImageRequest = remember {
        ImageRequest.Builder(context)
            .data("http://10.0.2.2:8080/images/gold.png?v=${System.currentTimeMillis()}")
            .error(R.drawable.gold)
            .placeholder(R.drawable.gold)
            .memoryCachePolicy(CachePolicy.DISABLED)
            .diskCachePolicy(CachePolicy.DISABLED)
            .build()
    }

    // Infinite animation for background sparkles and magical halo
    val infiniteTransition = rememberInfiniteTransition(label = "blessing_card_magic")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 0.94f,
        targetValue = 1.06f,
        animationSpec = infiniteRepeatable(
            animation = tween(1800, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "halo_pulse"
    )
    val starRotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(6000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "star_rotate"
    )

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            imageVector = Icons.Default.AutoAwesome,
                            contentDescription = null,
                            tint = GoldPrimary,
                            modifier = Modifier
                                .size(20.dp)
                                .rotate(starRotation)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "ได้ชื่อดีเป็นมงคล",
                            color = GoldLight,
                            fontWeight = FontWeight.Bold
                        )
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "กลับ",
                            tint = GoldLight
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = DarkBackground)
            )
        },
        containerColor = DarkBackground
    ) { paddingValues ->
        if (selectedItem == null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues),
                contentAlignment = Alignment.Center
            ) {
                Text("ไม่พบข้อมูลการ์ดมงคล", color = Color.White)
            }
        } else {
            val selectedName = selectedItem
            val fullName = snapshot.fullName(selectedName.thName)
            val goldCardText = Color(0xFF2B123A)
            val goldCardSecondaryText = Color(0xFF3B253F)
            val goldCardDivider = Color(0xFF8F6721).copy(alpha = 0.34f)

            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues)
            ) {
                // Background Ambient Glow & Floating Magic Sparkles
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(350.dp)
                        .background(
                            Brush.radialGradient(
                                colors = listOf(
                                    PurpleAccent.copy(alpha = 0.28f),
                                    GoldPrimary.copy(alpha = 0.15f),
                                    Color.Transparent
                                ),
                                center = Offset(400f, 100f),
                                radius = 900f
                            )
                        )
                )

                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState())
                        .padding(16.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Top
                ) {
                    // MAIN CAPTURABLE BLESSING CARD
                    BoxWithConstraints(
                        modifier = Modifier
                            .fillMaxWidth()
                            .aspectRatio(9f / 16f)
                            .capturable(captureController)
                            .clip(RoundedCornerShape(24.dp))
                            .background(
                                brush = Brush.verticalGradient(
                                    colors = listOf(
                                        Color(0xFFFFFBEB), // Soft Warm Sunlight Gold
                                        Color(0xFFFDE047), // Radiant Amber Gold
                                        Color(0xFFD97706), // Imperial Gold
                                        Color(0xFF78350F), // Deep Warm Bronze
                                        Color(0xFF38154C)  // Royal Purple Base Accent
                                    )
                                ),
                                shape = RoundedCornerShape(24.dp)
                            )
                            .border(
                                border = BorderStroke(
                                    width = 2.5.dp,
                                    brush = Brush.linearGradient(
                                        listOf(
                                            GoldPrimary,
                                            Color(0xFFF472B6),
                                            Color(0xFFA855F7),
                                            GoldPrimary
                                        )
                                    )
                                ),
                                shape = RoundedCornerShape(24.dp)
                            )
                            .padding(3.dp)
                            .border(
                                width = 1.dp,
                                color = Color(0xFFFFFDF0),
                                shape = RoundedCornerShape(21.dp)
                            )
                            .padding(14.dp)
                    ) {
                        // Background Radiant Halo behind Buddha Statue
                        Box(
                            modifier = Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = (buddhaImageTopPaddingFor(dayOfBirth) - 8.dp).coerceAtLeast(0.dp))
                                .size(
                                    width = maxWidth * 0.62f,
                                    height = maxHeight * 0.38f
                                )
                                .scale(pulseScale)
                                .background(
                                    Brush.radialGradient(
                                        colors = listOf(
                                            Color(0xFFFFFDF0).copy(alpha = 0.85f),
                                            Color(0xFFFDE68A).copy(alpha = 0.55f),
                                            Color(0xFFF59E0B).copy(alpha = 0.20f),
                                            Color.Transparent
                                        )
                                    )
                                )
                        )

                        // Floating Magic Sparkles inside Card
                        CardSparkleCanvas(
                            modifier = Modifier.matchParentSize(),
                            infiniteTransition = infiniteTransition
                        )

                        // Buddha Image for Day of Birth
                        AsyncImage(
                            model = buddhaImageRequest,
                            contentDescription = "พระประจำวัน${formatDayThai(dayOfBirth)}",
                            modifier = Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = buddhaImageTopPaddingFor(dayOfBirth))
                                .size(
                                    width = maxWidth * 0.52f,
                                    height = maxHeight * 0.36f
                                ),
                            contentScale = ContentScale.Fit,
                            alpha = 1f,
                            onSuccess = { buddhaImageUnavailable = false },
                            onError = {
                                if (!useLocalBuddhaImage) {
                                    useLocalBuddhaImage = true
                                } else {
                                    buddhaImageUnavailable = true
                                }
                            }
                        )

                        if (buddhaImageUnavailable) {
                            Text(
                                text = "ไม่พบภาพพระประจำวัน",
                                color = goldCardText.copy(alpha = 0.56f),
                                fontSize = 13.sp,
                                textAlign = TextAlign.Center,
                                modifier = Modifier.align(Alignment.TopCenter)
                            )
                        }

                        val nameFontSize = when {
                            fullName.length >= 26 -> 13.sp
                            fullName.length >= 21 -> 15.sp
                            fullName.length >= 16 -> 17.sp
                            fullName.length >= 11 -> 20.sp
                            else -> 24.sp
                        }

                        // INNER BLESSING CONTAINER (ORIGINAL CLEAN DESIGN)
                        Column(
                            modifier = Modifier
                                .align(Alignment.Center)
                                .fillMaxWidth(0.82f)
                                .widthIn(max = 310.dp)
                                .clip(RoundedCornerShape(22.dp))
                                .background(Color(0xFFFFF8DD).copy(alpha = 0.72f))
                                .border(
                                    width = 1.dp,
                                    color = Color(0xFFFFE88A).copy(alpha = 0.70f),
                                    shape = RoundedCornerShape(22.dp)
                                )
                                .padding(horizontal = 14.dp, vertical = 20.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            // Person Name Display
                            Text(
                                text = fullName,
                                color = goldCardText,
                                fontWeight = FontWeight.Bold,
                                fontSize = nameFontSize,
                                textAlign = TextAlign.Center,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis
                            )

                            Spacer(modifier = Modifier.height(12.dp))

                            // Meaning Display
                            Text(
                                text = selectedName.meaning.ifBlank { "ยังไม่มีข้อมูลความหมายของชื่อนี้ในฐานข้อมูล" },
                                color = goldCardSecondaryText,
                                fontWeight = FontWeight.Medium,
                                fontSize = 14.sp,
                                lineHeight = 20.sp,
                                textAlign = TextAlign.Center
                            )

                            HorizontalDivider(
                                modifier = Modifier.padding(vertical = 12.dp),
                                color = goldCardDivider
                            )

                            // Day-of-birth Kalakini safety information
                            Column(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalAlignment = Alignment.CenterHorizontally,
                                verticalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                Text(
                                    text = if (snapshot.policy.kalakiniScope == "not_applied") {
                                        "ผลลัพธ์นี้ไม่ได้ใช้ตัวกรองกาลกิณี"
                                    } else {
                                        "ผู้นี้เกิด${formatDayThai(dayOfBirth)} ไร้กาลกิณี"
                                    },
                                    color = Color(0xFF145A2B),
                                    fontWeight = FontWeight.SemiBold,
                                    fontSize = 13.sp,
                                    textAlign = TextAlign.Center
                                )

                                Text(
                                    text = "✦ ดีทั้งเลขศาสตร์และพลังเงา ✦",
                                    color = Color(0xFF8B5E14),
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 12.sp,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(top = 2.dp)
                                )
                            }

                            HorizontalDivider(
                                modifier = Modifier.padding(vertical = 12.dp),
                                color = goldCardDivider
                            )

                            val satDescDisplay = selectedName.satMiracleDesc.ifBlank { "ไม่มีคำอธิบายคู่เลขในฐานข้อมูล" }

                            BlessingNumberMeaningBlock(
                                pairsText = selectedName.satPairsDisplay,
                                pairTypesText = selectedName.satPairType,
                                description = satDescDisplay,
                                color = goldCardText,
                                onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                            )

                            val shaDescDisplay = selectedName.shaMiracleDesc.ifBlank { "ไม่มีคำอธิบายคู่เลขในฐานข้อมูล" }

                            BlessingNumberMeaningBlock(
                                pairsText = selectedName.shaPairsDisplay,
                                pairTypesText = selectedName.shaPairType,
                                description = shaDescDisplay,
                                color = goldCardSecondaryText,
                                onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                            )
                        }

                        // Decorative Frog Statue - Bottom Left (Overlapping text container for 3D depth)
                        AsyncImage(
                            model = frogLeftImageRequest,
                            contentDescription = "คางคกทองคำมงคลด้านซ้าย",
                            modifier = Modifier
                                .align(Alignment.BottomStart)
                                .offset(x = (-12).dp, y = 12.dp)
                                .size(
                                    width = maxWidth * 0.54f,
                                    height = maxHeight * 0.32f
                                ),
                            contentScale = ContentScale.Fit
                        )

                        // Decorative Gold & Treasure Image - Bottom Right (Overlapping text container for 3D depth)
                        AsyncImage(
                            model = goldRightImageRequest,
                            contentDescription = "กองเงินกองทองด้านขวา",
                            modifier = Modifier
                                .align(Alignment.BottomEnd)
                                .offset(x = 12.dp, y = 12.dp)
                                .size(
                                    width = maxWidth * 0.54f,
                                    height = maxHeight * 0.32f
                                ),
                            contentScale = ContentScale.Fit
                        )
                    }



                    Spacer(modifier = Modifier.height(16.dp))

                    // SHARE & SAVE BUTTON
                    Button(
                        onClick = {
                            scope.launch {
                                try {
                                    val bitmap = captureController.captureAsync().await()
                                    val success = ImageUtils.saveBitmapToGallery(context, bitmap.asAndroidBitmap())
                                    if (success) {
                                        Toast.makeText(context, "บันทึกรูปลงเครื่องสำเร็จ!", Toast.LENGTH_SHORT).show()
                                    } else {
                                        Toast.makeText(context, "เกิดข้อผิดพลาดในการบันทึกรูปภาพ", Toast.LENGTH_SHORT).show()
                                    }
                                } catch (_: Throwable) {
                                    Toast.makeText(context, "ไม่สามารถจับภาพหน้าจอได้", Toast.LENGTH_SHORT).show()
                                }
                            }
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(52.dp),
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = Color.Unspecified),
                        contentPadding = PaddingValues()
                    ) {
                        Box(
                            modifier = Modifier
                                .fillMaxSize()
                                .background(
                                    Brush.horizontalGradient(
                                        listOf(
                                            GoldPrimary,
                                            Color(0xFFF59E0B),
                                            GoldPrimary
                                        )
                                    ),
                                    shape = RoundedCornerShape(14.dp)
                                ),
                            contentAlignment = Alignment.Center
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Icon(
                                    imageVector = Icons.Default.Share,
                                    contentDescription = null,
                                    tint = Color.Black
                                )
                                Spacer(modifier = Modifier.width(10.dp))
                                Text(
                                    text = "บันทึกภาพการ์ดชื่อดี",
                                    color = Color.Black,
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 15.sp
                                )
                            }
                        }
                    }

                    Spacer(modifier = Modifier.height(24.dp))
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
private fun CardSparkleCanvas(modifier: Modifier, infiniteTransition: InfiniteTransition) {
    val alpha1 by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.9f,
        animationSpec = infiniteRepeatable(
            animation = tween(1400, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "card_sparkle_1"
    )
    val alpha2 by infiniteTransition.animateFloat(
        initialValue = 0.85f,
        targetValue = 0.2f,
        animationSpec = infiniteRepeatable(
            animation = tween(1900, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "card_sparkle_2"
    )

    Canvas(modifier = modifier) {
        // Floating Star Dots on Gold Background
        drawCircle(
            color = Color(0xFFFFFDF0).copy(alpha = alpha1),
            radius = 3.5.dp.toPx(),
            center = Offset(size.width * 0.15f, size.height * 0.12f)
        )
        drawCircle(
            color = Color(0xFFFFFDF0).copy(alpha = alpha2),
            radius = 4.dp.toPx(),
            center = Offset(size.width * 0.85f, size.height * 0.18f)
        )
        drawCircle(
            color = Color(0xFFFDE68A).copy(alpha = alpha1),
            radius = 3.dp.toPx(),
            center = Offset(size.width * 0.10f, size.height * 0.38f)
        )
        drawCircle(
            color = Color(0xFFFDE68A).copy(alpha = alpha2),
            radius = 3.5.dp.toPx(),
            center = Offset(size.width * 0.90f, size.height * 0.42f)
        )
    }
}



private fun formatDayThai(day: String): String {
    return when (day.lowercase()) {
        "sunday" -> "วันอาทิตย์"
        "monday" -> "วันจันทร์"
        "tuesday" -> "วันอังคาร"
        "wednesday1" -> "วันพุธ (กลางวัน)"
        "wednesday2" -> "วันพุธ (กลางคืน)"
        "wednesday" -> "วันพุธ (กลางวัน)"
        "wednesday_night" -> "วันพุธ (กลางคืน)"
        "thursday" -> "วันพฤหัสบดี"
        "friday" -> "วันศุกร์"
        "saturday" -> "วันเสาร์"
        else -> day
    }
}

@Composable
private fun BlessingNumberMeaningBlock(
    label: String? = null,
    pairsText: String,
    pairTypesText: String,
    description: String,
    color: Color,
    onPairClick: ((String) -> Unit)? = null
) {
    val descriptionLines = remember(description) { splitBlessingDescriptions(description) }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        if (label != null) {
            Text(
                text = "• $label",
                color = color,
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold,
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(3.dp))
        }
        if (pairsText.isNotBlank()) {
            com.miracle.names.ui.results.PairCircles(
                pairsStr = pairsText,
                typesStr = pairTypesText,
                onPairClick = onPairClick
            )
        } else {
            Text(
                text = "ไม่มีข้อมูลคู่เลข",
                color = color.copy(alpha = 0.72f),
                fontSize = 11.sp
            )
        }

        Spacer(modifier = Modifier.height(4.dp))

        Column(
            modifier = Modifier.fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(2.dp)
        ) {
            descriptionLines.forEach { line ->
                Text(
                    text = line,
                    color = color,
                    fontSize = 12.sp,
                    textAlign = TextAlign.Center,
                    lineHeight = 17.sp,
                    modifier = Modifier.fillMaxWidth()
                )
            }
        }
    }
}

private fun splitBlessingDescriptions(description: String): List<String> {
    return description
        .split(Regex("\\s*(?:\\n|/|•|,|，|、)\\s*"))
        .map { it.trim() }
        .filter { it.isNotEmpty() }
        .ifEmpty { listOf(description.trim()) }
}

private fun buddhaImageFileNameFor(day: String): String {
    return when (day.lowercase()) {
        "wednesday" -> "wednesday1_buddha.png"
        "wednesday_night" -> "wednesday2_buddha.png"
        else -> "${day.lowercase()}_buddha.png"
    }
}

private fun localBuddhaPngResourceFor(day: String): Int {
    return when (day.lowercase()) {
        "sunday" -> R.drawable.sunday_buddha
        "monday" -> R.drawable.monday_buddha
        "tuesday" -> R.drawable.tuesday_buddha
        "wednesday1" -> R.drawable.wednesday1_buddha
        "wednesday2" -> R.drawable.wednesday2_buddha
        "wednesday" -> R.drawable.wednesday1_buddha
        "wednesday_night" -> R.drawable.wednesday2_buddha
        "thursday" -> R.drawable.thursday_buddha
        "friday" -> R.drawable.friday_buddha
        "saturday" -> R.drawable.saturday_buddha
        else -> R.drawable.sunday_buddha
    }
}

private fun buddhaImageTopPaddingFor(day: String): androidx.compose.ui.unit.Dp {
    return when (day.lowercase()) {
        "tuesday" -> 4.dp
        else -> 18.dp
    }
}
