package com.miracle.names.ui.blessing

import android.app.Activity
import android.widget.Toast
import androidx.compose.animation.core.*
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.Image
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
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asAndroidBitmap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
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
import com.miracle.names.data.remote.MiracleApiConfig
import com.miracle.names.ui.theme.*
import com.miracle.names.utils.ImageUtils
import com.miracle.names.viewmodel.AppliedSearchSnapshot
import com.miracle.names.viewmodel.NameSearchViewModel
import dev.shreyaspatil.capturable.capturable
import dev.shreyaspatil.capturable.controller.rememberCaptureController
import kotlinx.coroutines.launch
import java.util.Locale
import androidx.compose.runtime.ExperimentalComposeApi
import androidx.compose.material.icons.filled.BugReport
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import kotlin.math.roundToInt

@OptIn(ExperimentalMaterial3Api::class, ExperimentalComposeUiApi::class, ExperimentalComposeApi::class)
@Composable
fun BlessingCardScreen(
    item: NameResultItem? = null,
    snapshot: AppliedSearchSnapshot,
    viewModel: NameSearchViewModel,
    onBackClick: () -> Unit = { viewModel.selectNameForBlessingCard(null) },
    onOpenDebugPurchase: (() -> Unit)? = null
) {
    val selectedItem = item ?: viewModel.selectedItem.collectAsState().value
    val dayOfBirth = snapshot.dayOfBirth
    val context = LocalContext.current
    val isUnlocked by viewModel.isUnlocked.collectAsState()
    val scope = rememberCoroutineScope()
    val captureController = rememberCaptureController()
    var useLocalBuddhaImage by remember(dayOfBirth) { mutableStateOf(false) }
    var buddhaImageUnavailable by remember(dayOfBirth) { mutableStateOf(false) }
    val localBuddhaImageResId = remember(dayOfBirth) { localBuddhaPngResourceFor(dayOfBirth) }
    val buddhaImageData = if (useLocalBuddhaImage) {
        localBuddhaImageResId
    } else {
        "${MiracleApiConfig.assetUrl("buddhas/${buddhaImageFileNameFor(dayOfBirth)}")}?v=${System.currentTimeMillis()}"
    }
    val buddhaImageRequest = remember(dayOfBirth, useLocalBuddhaImage) {
        ImageRequest.Builder(context)
            .data(buddhaImageData)
            .memoryCachePolicy(CachePolicy.DISABLED)
            .diskCachePolicy(CachePolicy.DISABLED)
            .build()
    }
    var buddhaImageLoaded by remember(dayOfBirth, useLocalBuddhaImage) { mutableStateOf(false) }

    var selectedTheme by remember { mutableStateOf(WallpaperTheme.EMERALD) }

    val wallpaperBgRequest = remember(selectedTheme) {
        ImageRequest.Builder(context)
            .data("${MiracleApiConfig.assetUrl("images/${selectedTheme.bgFileName}")}?v=${System.currentTimeMillis()}")
            .error(selectedTheme.fallbackResId)
            .placeholder(selectedTheme.fallbackResId)
            .memoryCachePolicy(CachePolicy.DISABLED)
            .diskCachePolicy(CachePolicy.DISABLED)
            .build()
    }

    val frogLeftImageRequest = remember {
        ImageRequest.Builder(context)
            .data("${MiracleApiConfig.assetUrl("images/frog_left.png")}?v=${System.currentTimeMillis()}")
            .error(R.drawable.frog_left)
            .placeholder(R.drawable.frog_left)
            .memoryCachePolicy(CachePolicy.DISABLED)
            .diskCachePolicy(CachePolicy.DISABLED)
            .build()
    }

    val goldRightImageRequest = remember {
        ImageRequest.Builder(context)
            .data("${MiracleApiConfig.assetUrl("images/gold.png")}?v=${System.currentTimeMillis()}")
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
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Image(
                            painter = painterResource(id = R.drawable.app_logo),
                            contentDescription = "โลโก้ชื่อดี",
                            modifier = Modifier
                                .size(30.dp)
                                .clip(CircleShape)
                                .border(1.dp, GoldPrimary.copy(alpha = 0.6f), CircleShape)
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
                actions = {
                    onOpenDebugPurchase?.let { openDebug ->
                        IconButton(onClick = openDebug) {
                            Icon(
                                imageVector = Icons.Default.BugReport,
                                contentDescription = "เมนู Debug purchase",
                                tint = GoldLight
                            )
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = GreenHeader)
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
            val rawFullName = snapshot.fullName(selectedName.thName)
            val fullName = viewModel.getDisplayName(rawFullName)
            val goldCardText = selectedTheme.nameTextColor
            val goldCardSecondaryText = selectedTheme.meaningTextColor
            val goldCardDivider = selectedTheme.dividerColor

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
                                    GreenHeader.copy(alpha = 0.38f),
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
                    if (!isUnlocked) {
                        Surface(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(bottom = 16.dp),
                            shape = RoundedCornerShape(14.dp),
                            color = GoldPrimary,
                            contentColor = Color(0xFF123322),
                            shadowElevation = 5.dp
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 14.dp, vertical = 11.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(9.dp)
                            ) {
                                Text("🔒", fontSize = 16.sp)
                                Text(
                                    text = "ตัวอย่างการ์ด · เปิดชื่อทั้งหมดเพื่อสร้างแรงดึงดูด",
                                    modifier = Modifier.weight(1f),
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.SemiBold,
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis
                                )
                                TextButton(
                                    onClick = { viewModel.unlockAllNames(context as? Activity) },
                                    colors = ButtonDefaults.textButtonColors(contentColor = Color(0xFF123322)),
                                    contentPadding = PaddingValues(horizontal = 8.dp, vertical = 4.dp)
                                ) {
                                    Text("เปิดชื่อ", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                                }
                            }
                        }
                    }

                    val nameFontSize = remember(fullName) {
                        when {
                            fullName.length >= 26 -> 13.sp
                            fullName.length >= 21 -> 15.sp
                            fullName.length >= 16 -> 17.sp
                            fullName.length >= 11 -> 19.sp
                            else -> 23.sp
                        }
                    }

                    // THEME SELECTOR PILLS
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 12.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        WallpaperTheme.values().forEach { theme ->
                            val isSelected = theme == selectedTheme
                            Surface(
                                onClick = { selectedTheme = theme },
                                shape = RoundedCornerShape(20.dp),
                                color = if (isSelected) GoldPrimary else CardBgDark,
                                contentColor = if (isSelected) Color(0xFF1B1405) else GoldLight,
                                border = BorderStroke(
                                    1.dp,
                                    if (isSelected) GoldLight else GoldPrimary.copy(alpha = 0.35f)
                                ),
                                modifier = Modifier.weight(1f)
                            ) {
                                Box(
                                    modifier = Modifier.padding(vertical = 8.dp, horizontal = 4.dp),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Text(
                                        text = theme.title,
                                        fontSize = 11.5.sp,
                                        fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                                        maxLines = 1
                                    )
                                }
                            }
                        }
                    }

                    // MAIN CAPTURABLE BLESSING CARD / WALLPAPER (Optimized for 9:19.5 smartphone screens)
                    BoxWithConstraints(
                        modifier = Modifier
                            .fillMaxWidth()
                            .aspectRatio(9f / 19.5f)
                            .capturable(captureController)
                            .background(Color(0xFFFBF8F0))
                    ) {
                        // Dynamic Wallpaper Background (served from Go backend)
                        AsyncImage(
                            model = wallpaperBgRequest,
                            contentDescription = "ภาพพื้นหลังวอลเปเปอร์มงคล",
                            modifier = Modifier.matchParentSize(),
                            contentScale = ContentScale.Crop
                        )

                        val isTuesday = remember(dayOfBirth) { dayOfBirth.equals("tuesday", ignoreCase = true) }
                        val tuesdayBoxSize = maxWidth * 0.90f
                        val tuesdayYOffset = maxHeight * 0.32f - (tuesdayBoxSize * 0.655f)

                        val haloWidth = if (isTuesday) maxWidth * 0.88f else maxWidth * 0.62f
                        val haloHeight = if (isTuesday) (tuesdayBoxSize * 0.50f) else maxHeight * 0.28f

                        // Background Radiant Halo behind Buddha Statue
                        Box(
                            modifier = (if (isTuesday) {
                                Modifier
                                    .align(Alignment.TopCenter)
                                    .offset(y = tuesdayYOffset + (tuesdayBoxSize * 0.25f))
                                    .size(width = haloWidth, height = haloHeight)
                            } else {
                                Modifier
                                    .align(Alignment.TopCenter)
                                    .padding(top = maxHeight * 0.05f)
                                    .size(width = haloWidth, height = haloHeight)
                            })
                                .scale(pulseScale)
                                .background(
                                    Brush.radialGradient(
                                        colors = selectedTheme.haloColors
                                    )
                                )
                        )

                        // Floating Magic Sparkles inside Card
                        CardSparkleCanvas(
                            modifier = Modifier.matchParentSize(),
                            infiniteTransition = infiniteTransition
                        )

                        val isStandingBuddha = remember(dayOfBirth) {
                            dayOfBirth.equals("sunday", ignoreCase = true) ||
                            dayOfBirth.equals("monday", ignoreCase = true) ||
                            dayOfBirth.equals("friday", ignoreCase = true)
                        }

                        val buddhaImageModifier = if (isTuesday) {
                            Modifier
                                .align(Alignment.TopCenter)
                                .offset(y = tuesdayYOffset)
                                .size(tuesdayBoxSize)
                        } else {
                            Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = if (isStandingBuddha) maxHeight * 0.035f else maxHeight * 0.05f)
                                .size(
                                    width = if (isStandingBuddha) maxWidth * 0.54f else maxWidth * 0.52f,
                                    height = if (isStandingBuddha) maxHeight * 0.28f else maxHeight * 0.27f
                                )
                        }

                        if (!buddhaImageLoaded && !buddhaImageUnavailable) {
                            BuddhaLightCirclePlaceholder(
                                modifier = buddhaImageModifier,
                                infiniteTransition = infiniteTransition
                            )
                        }

                        // Buddha Image for Day of Birth
                        val buddhaImageAlpha by animateFloatAsState(
                            targetValue = if (buddhaImageLoaded) 1f else 0f,
                            animationSpec = tween(520, easing = FastOutSlowInEasing),
                            label = "buddha_image_fade_in"
                        )
                        AsyncImage(
                            model = buddhaImageRequest,
                            contentDescription = "พระประจำวัน${formatDayThai(dayOfBirth)}",
                            modifier = buddhaImageModifier,
                            contentScale = ContentScale.Fit,
                            alignment = if (isTuesday) Alignment.BottomCenter else Alignment.Center,
                            alpha = buddhaImageAlpha,
                            onLoading = { buddhaImageLoaded = false },
                            onSuccess = {
                                buddhaImageLoaded = true
                                buddhaImageUnavailable = false
                            },
                            onError = {
                                buddhaImageLoaded = false
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
                            fullName.length >= 11 -> 19.sp
                            else -> 23.sp
                        }
                        val blessingOrnamentWidth = maxWidth * 0.25f
                        val blessingOrnamentHeight = maxHeight * 0.105f

                        // INNER BLESSING CONTAINER (POSITIONED DIRECTLY UNDERNEATH THE BUDDHA STATUE)
                        Column(
                            modifier = Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = maxHeight * 0.32f)
                                .fillMaxWidth(0.88f)
                                .widthIn(max = 340.dp)
                                .shadow(
                                    elevation = if (selectedTheme == WallpaperTheme.IVORY) 8.dp else 4.dp,
                                    shape = RoundedCornerShape(22.dp),
                                    ambientColor = Color(0x333D270C),
                                    spotColor = Color(0x333D270C)
                                )
                                .clip(RoundedCornerShape(22.dp))
                                .background(Brush.verticalGradient(selectedTheme.cardBackgroundColors))
                                .border(
                                    width = 1.dp,
                                    brush = Brush.verticalGradient(selectedTheme.cardBorderColors),
                                    shape = RoundedCornerShape(22.dp)
                                )
                                .padding(horizontal = 14.dp, vertical = 12.dp),
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

                            Spacer(modifier = Modifier.height(6.dp))

                            // Meaning Display
                            Text(
                                text = selectedName.meaning.ifBlank { "ยังไม่มีข้อมูลความหมายของชื่อนี้ในฐานข้อมูล" },
                                color = goldCardSecondaryText,
                                fontWeight = FontWeight.Medium,
                                fontSize = 13.sp,
                                lineHeight = 18.sp,
                                textAlign = TextAlign.Center
                            )

                            HorizontalDivider(
                                modifier = Modifier.padding(vertical = 7.dp),
                                color = goldCardDivider
                            )

                            // Day-of-birth Kalakini safety information
                            Column(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalAlignment = Alignment.CenterHorizontally,
                                verticalArrangement = Arrangement.spacedBy(2.dp)
                            ) {
                                Text(
                                    text = if (snapshot.policy.kalakiniScope == "not_applied") {
                                        "ผลลัพธ์นี้ไม่ได้ใช้ตัวกรองกาลกิณี"
                                    } else {
                                        "ผู้นี้เกิด${formatDayThai(dayOfBirth)} ไร้กาลกิณี"
                                    },
                                    color = selectedTheme.kalakiniTextColor,
                                    fontWeight = FontWeight.SemiBold,
                                    fontSize = 12.5.sp,
                                    textAlign = TextAlign.Center
                                )

                                Text(
                                    text = "✦ เลขศาสตร์นำพา พลังเงาเกื้อหนุน ✦",
                                    color = selectedTheme.quoteTextColor,
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 11.5.sp,
                                    textAlign = TextAlign.Center
                                )
                            }

                            HorizontalDivider(
                                modifier = Modifier.padding(vertical = 7.dp),
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

                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(top = 4.dp),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.Bottom
                            ) {
                                AsyncImage(
                                    model = frogLeftImageRequest,
                                    contentDescription = "คางคกทองคำมงคลด้านซ้าย",
                                    modifier = Modifier.size(
                                        width = blessingOrnamentWidth,
                                        height = blessingOrnamentHeight
                                    ),
                                    contentScale = ContentScale.Fit,
                                    alpha = 0.88f
                                )

                                AsyncImage(
                                    model = goldRightImageRequest,
                                    contentDescription = "กองเงินกองทองด้านขวา",
                                    modifier = Modifier.size(
                                        width = blessingOrnamentWidth,
                                        height = blessingOrnamentHeight
                                    ),
                                    contentScale = ContentScale.Fit,
                                    alpha = 0.88f
                                )
                            }
                        }

                        // Brand footer
                        Box(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .padding(bottom = 14.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = "ชื่อดี.com",
                                color = selectedTheme.footerTextColor,
                                fontSize = 11.sp,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(20.dp))

                    var showSaveOptionDialog by remember { mutableStateOf(false) }
                    val socialCaptureController = rememberCaptureController()

                    // SAVE & WALLPAPER ACTION BUTTON
                    Button(
                        onClick = { showSaveOptionDialog = true },
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
                                    text = "บันทึกภาพ/แชร์",
                                    color = Color.Black,
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 15.sp
                                )
                            }
                        }
                    }

                    // Off-screen 1:1 Social Blessing Card for Instagram / Facebook / Social Sharing
                    val socialCanvasSize = 420.dp
                    Box(
                        modifier = Modifier
                            .offset(x = 6000.dp)
                            .requiredSize(socialCanvasSize)
                            .capturable(socialCaptureController)
                            .background(Color(0xFFFBF8F0))
                    ) {
                        // 1:1 Social Card Content
                        AsyncImage(
                            model = wallpaperBgRequest,
                            contentDescription = null,
                            modifier = Modifier.matchParentSize(),
                            contentScale = ContentScale.Crop
                        )

                        val isTuesday = remember(dayOfBirth) { dayOfBirth.equals("tuesday", ignoreCase = true) }
                        val isStandingBuddha = remember(dayOfBirth) {
                            dayOfBirth.equals("sunday", ignoreCase = true) ||
                            dayOfBirth.equals("monday", ignoreCase = true) ||
                            dayOfBirth.equals("friday", ignoreCase = true)
                        }

                        val sTuesdayBoxSize = socialCanvasSize * 0.58f
                        val sTuesdayYOffset = socialCanvasSize * 0.25f - (sTuesdayBoxSize * 0.655f)
                        val sBuddhaWidth = if (isTuesday) sTuesdayBoxSize else (if (isStandingBuddha) socialCanvasSize * 0.35f else socialCanvasSize * 0.33f)
                        val sBuddhaHeight = if (isTuesday) sTuesdayBoxSize else (if (isStandingBuddha) socialCanvasSize * 0.25f else socialCanvasSize * 0.23f)
                        val sBuddhaTopPadding = if (isTuesday) sTuesdayYOffset else (if (isStandingBuddha) socialCanvasSize * 0.015f else socialCanvasSize * 0.022f)

                        // Top Buddha Statue
                        AsyncImage(
                            model = buddhaImageRequest,
                            contentDescription = null,
                            modifier = Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = sBuddhaTopPadding)
                                .size(width = sBuddhaWidth, height = sBuddhaHeight),
                            contentScale = ContentScale.Fit,
                            alignment = if (isTuesday) Alignment.BottomCenter else Alignment.Center
                        )

                        // Inner Card
                        Column(
                            modifier = Modifier
                                .align(Alignment.TopCenter)
                                .padding(top = socialCanvasSize * 0.245f)
                                .fillMaxWidth(0.91f)
                                .height(socialCanvasSize * 0.73f)
                                .shadow(
                                    elevation = if (selectedTheme == WallpaperTheme.IVORY) 6.dp else 3.dp,
                                    shape = RoundedCornerShape(18.dp),
                                    ambientColor = Color(0x333D270C),
                                    spotColor = Color(0x333D270C)
                                )
                                .clip(RoundedCornerShape(18.dp))
                                .background(Brush.verticalGradient(selectedTheme.cardBackgroundColors))
                                .border(
                                    width = 1.2.dp,
                                    brush = Brush.verticalGradient(selectedTheme.cardBorderColors),
                                    shape = RoundedCornerShape(18.dp)
                                )
                                .padding(horizontal = 14.dp, vertical = 2.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.SpaceBetween
                        ) {
                            Column(
                                horizontalAlignment = Alignment.CenterHorizontally,
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text(
                                    text = fullName,
                                    color = goldCardText,
                                    fontWeight = FontWeight.Bold,
                                    fontSize = nameFontSize,
                                    textAlign = TextAlign.Center,
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis,
                                    modifier = Modifier.padding(top = 14.dp)
                                )

                                Text(
                                    text = selectedName.meaning.ifBlank { "ยังไม่มีข้อมูลความหมายของชื่อนี้ในฐานข้อมูล" },
                                    color = goldCardSecondaryText,
                                    fontSize = 11.5.sp,
                                    lineHeight = 15.sp,
                                    fontWeight = FontWeight.Medium,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(top = 3.dp, start = 4.dp, end = 4.dp)
                                )

                                HorizontalDivider(
                                    modifier = Modifier.padding(vertical = 2.5.dp),
                                    color = goldCardDivider
                                )

                                Text(
                                    text = "ผู้นี้เกิด${formatDayThai(dayOfBirth)} ไร้กาลกิณี",
                                    color = selectedTheme.kalakiniTextColor,
                                    fontWeight = FontWeight.SemiBold,
                                    fontSize = 11.sp,
                                    textAlign = TextAlign.Center
                                )

                                Text(
                                    text = "✦ เลขศาสตร์นำพา พลังเงาเกื้อหนุน ✦",
                                    color = selectedTheme.quoteTextColor,
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 10.sp,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(top = 1.dp)
                                )

                                HorizontalDivider(
                                    modifier = Modifier.padding(vertical = 2.5.dp),
                                    color = goldCardDivider
                                )

                                Column(
                                    verticalArrangement = Arrangement.spacedBy(3.dp),
                                    horizontalAlignment = Alignment.CenterHorizontally,
                                    modifier = Modifier.fillMaxWidth()
                                ) {
                                    BlessingNumberMeaningBlock(
                                        pairsText = selectedName.satPairsDisplay,
                                        pairTypesText = selectedName.satPairType,
                                        description = selectedName.satMiracleDesc.ifBlank { "ไม่มีคำอธิบายคู่เลขในฐานข้อมูล" },
                                        color = goldCardText,
                                        circleSize = 30.dp,
                                        fontSize = 13.sp,
                                        descFontSize = 11.5.sp,
                                        strokeWidth = 1.5.dp,
                                        onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                                    )

                                    BlessingNumberMeaningBlock(
                                        pairsText = selectedName.shaPairsDisplay,
                                        pairTypesText = selectedName.shaPairType,
                                        description = selectedName.shaMiracleDesc.ifBlank { "ไม่มีคำอธิบายคู่เลขในฐานข้อมูล" },
                                        color = goldCardSecondaryText,
                                        circleSize = 30.dp,
                                        fontSize = 13.sp,
                                        descFontSize = 11.5.sp,
                                        strokeWidth = 1.5.dp,
                                        onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                                    )
                                }
                            }

                            // Bottom Frog, Gold & Brand
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(start = 4.dp, end = 4.dp, bottom = 4.dp),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                AsyncImage(
                                    model = frogLeftImageRequest,
                                    contentDescription = "คางคกทองคำมงคลด้านซ้าย",
                                    modifier = Modifier.size(width = 68.dp, height = 44.dp),
                                    contentScale = ContentScale.Fit,
                                    alpha = 0.95f
                                )

                                Text(
                                    text = "ชื่อดี.com",
                                    color = selectedTheme.footerTextColor,
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.Bold
                                )

                                AsyncImage(
                                    model = goldRightImageRequest,
                                    contentDescription = "กองเงินกองทองด้านขวา",
                                    modifier = Modifier.size(width = 68.dp, height = 44.dp),
                                    contentScale = ContentScale.Fit,
                                    alpha = 0.95f
                                )
                            }
                        }
                    }

                    if (showSaveOptionDialog) {
                        AlertDialog(
                            onDismissRequest = { showSaveOptionDialog = false },
                            title = {
                                Text(
                                    text = "เลือกวิธีใช้การ์ดมงคล",
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 17.sp,
                                    color = MaterialTheme.colorScheme.onSurface
                                )
                            },
                            text = {
                                Column(
                                    verticalArrangement = Arrangement.spacedBy(10.dp),
                                    modifier = Modifier.fillMaxWidth()
                                ) {
                                    Text(
                                        text = "คุณสามารถเก็บภาพไว้ใน Photos หรือแชร์ผ่านแอปที่ติดตั้งอยู่",
                                        fontSize = 14.sp,
                                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                                    )
                                    
                                    Button(
                                        onClick = {
                                            showSaveOptionDialog = false
                                            scope.launch {
                                                try {
                                                    val bitmap = captureController.captureAsync().await()
                                                    val androidBitmap = bitmap.asAndroidBitmap()
                                                    val saved = ImageUtils.saveBitmapToGallery(context, androidBitmap)
                                                    if (saved) {
                                                        Toast.makeText(context, "บันทึกรูปลงคลังภาพสำเร็จ!", Toast.LENGTH_SHORT).show()
                                                    } else {
                                                        Toast.makeText(context, "เกิดข้อผิดพลาดในการบันทึกรูปภาพ", Toast.LENGTH_SHORT).show()
                                                    }
                                                } catch (_: Throwable) {
                                                    Toast.makeText(context, "ไม่สามารถจับภาพได้", Toast.LENGTH_SHORT).show()
                                                }
                                            }
                                        },
                                        colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary),
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .height(48.dp),
                                        shape = RoundedCornerShape(12.dp)
                                    ) {
                                        Text(
                                            text = "🖼️  บันทึกเป็นรูปภาพ",
                                            color = Color.Black,
                                            fontWeight = FontWeight.Bold,
                                            fontSize = 15.sp
                                        )
                                    }

                                    Button(
                                        onClick = {
                                            showSaveOptionDialog = false
                                            scope.launch {
                                                try {
                                                    val bitmap = socialCaptureController.captureAsync().await()
                                                    val cardBitmap = bitmap.asAndroidBitmap()
                                                    ImageUtils.shareImage(context, cardBitmap)
                                                } catch (_: Throwable) {
                                                    Toast.makeText(context, "ไม่สามารถแชร์การ์ดรูปภาพได้", Toast.LENGTH_SHORT).show()
                                                }
                                            }
                                        },
                                        colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF1877F2)),
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .height(48.dp),
                                        shape = RoundedCornerShape(12.dp)
                                    ) {
                                        Text(
                                            text = "📲  แชร์ไปยังโซเชียล",
                                            color = Color.White,
                                            fontWeight = FontWeight.Bold,
                                            fontSize = 15.sp
                                        )
                                    }
                                }
                            },
                            confirmButton = {},
                            dismissButton = {
                                TextButton(onClick = { showSaveOptionDialog = false }) {
                                    Text("ยกเลิก", color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f), fontSize = 14.sp)
                                }
                            }
                        )
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

@Composable
private fun BuddhaLightCirclePlaceholder(
    modifier: Modifier,
    infiniteTransition: InfiniteTransition
) {
    val glowAlpha by infiniteTransition.animateFloat(
        initialValue = 0.32f,
        targetValue = 0.90f,
        animationSpec = infiniteRepeatable(
            animation = tween(1500, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "buddha_light_circle_glow"
    )
    val ringScale by infiniteTransition.animateFloat(
        initialValue = 0.72f,
        targetValue = 1.18f,
        animationSpec = infiniteRepeatable(
            animation = tween(2300, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "buddha_light_circle_ring"
    )

    Canvas(modifier = modifier) {
        val minSide = size.minDimension
        val center = Offset(size.width * 0.5f, size.height * 0.44f)
        val gold = Color(0xFFFFF3B0)
        val softWhite = Color(0xFFFFFDF0)
        val amber = Color(0xFFF59E0B)

        drawCircle(
            color = amber.copy(alpha = glowAlpha * 0.12f),
            radius = minSide * 0.54f,
            center = center
        )
        drawCircle(
            color = gold.copy(alpha = glowAlpha * 0.24f),
            radius = minSide * 0.40f,
            center = center
        )
        drawCircle(
            color = softWhite.copy(alpha = glowAlpha * 0.32f),
            radius = minSide * 0.22f,
            center = center
        )
        drawCircle(
            color = gold.copy(alpha = glowAlpha * 0.42f),
            radius = minSide * 0.42f * ringScale,
            center = center,
            style = Stroke(width = 2.4.dp.toPx())
        )
        drawCircle(
            color = softWhite.copy(alpha = glowAlpha * 0.78f),
            radius = minSide * 0.030f,
            center = Offset(size.width * 0.50f, size.height * 0.44f)
        )
        drawCircle(
            color = softWhite.copy(alpha = glowAlpha * 0.58f),
            radius = minSide * 0.016f,
            center = Offset(size.width * 0.34f, size.height * 0.25f)
        )
        drawCircle(
            color = gold.copy(alpha = glowAlpha * 0.62f),
            radius = minSide * 0.014f,
            center = Offset(size.width * 0.68f, size.height * 0.32f)
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
    circleSize: androidx.compose.ui.unit.Dp = 36.dp,
    fontSize: androidx.compose.ui.unit.TextUnit = 14.sp,
    descFontSize: androidx.compose.ui.unit.TextUnit = 12.sp,
    strokeWidth: androidx.compose.ui.unit.Dp = 2.dp,
    onPairClick: ((String) -> Unit)? = null
) {
    val descriptionLines = remember(description) { splitBlessingDescriptions(description) }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 2.dp),
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
                circleSize = circleSize,
                fontSize = fontSize,
                strokeWidth = strokeWidth,
                spacing = 6.dp,
                onPairClick = onPairClick
            )
        } else {
            Text(
                text = "ไม่มีข้อมูลคู่เลข",
                color = color.copy(alpha = 0.72f),
                fontSize = 11.sp
            )
        }

        Spacer(modifier = Modifier.height(3.dp))

        Column(
            modifier = Modifier.fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(1.dp)
        ) {
            descriptionLines.forEach { line ->
                Text(
                    text = line,
                    color = color,
                    fontSize = descFontSize,
                    textAlign = TextAlign.Center,
                    lineHeight = 16.sp,
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

enum class WallpaperTheme(
    val id: String,
    val title: String,
    val bgFileName: String,
    val fallbackResId: Int,
    val cardBackground: Color,
    val cardBackgroundColors: List<Color>,
    val cardBorderColors: List<Color>,
    val nameTextColor: Color,
    val meaningTextColor: Color,
    val kalakiniTextColor: Color,
    val quoteTextColor: Color,
    val dividerColor: Color,
    val footerTextColor: Color,
    val outerBorderColors: List<Color>,
    val haloColors: List<Color>
) {
    IVORY(
        id = "ivory",
        title = "✨ ทองคำ",
        bgFileName = "blessing_wallpaper_bg_ivory.jpg",
        fallbackResId = R.drawable.blessing_wallpaper_bg_ivory,
        cardBackground = Color(0xFFF3E4C2).copy(alpha = 0.95f),
        cardBackgroundColors = listOf(
            Color(0xFFF9F0D9).copy(alpha = 0.95f),
            Color(0xFFEBD8AA).copy(alpha = 0.95f)
        ),
        cardBorderColors = listOf(
            Color(0xFFF7E5BC),
            Color(0xFFD4AF37).copy(alpha = 0.85f),
            Color(0xFFE8D3A2)
        ),
        nameTextColor = Color(0xFF24170A),
        meaningTextColor = Color(0xFF503D2A),
        kalakiniTextColor = Color(0xFF1E582D),
        quoteTextColor = Color(0xFF7A5012),
        dividerColor = Color(0xFFC4A255).copy(alpha = 0.45f),
        footerTextColor = Color(0xFF6B4918),
        outerBorderColors = listOf(Color(0xFFFFF8DE), Color(0xFFDFB86C), Color(0xFFFFF8DE)),
        haloColors = listOf(
            Color(0xFFFFF2D0).copy(alpha = 0.78f),
            Color(0xFFF6D88E).copy(alpha = 0.45f),
            Color(0xFFE2AE48).copy(alpha = 0.18f),
            Color.Transparent
        )
    ),
    EMERALD(
        id = "emerald",
        title = "🌿 เขียวมรกต",
        bgFileName = "blessing_wallpaper_bg_emerald.jpg",
        fallbackResId = R.drawable.blessing_wallpaper_bg_emerald,
        cardBackground = Color(0xFF072718).copy(alpha = 0.88f),
        cardBackgroundColors = listOf(
            Color(0xFF072718).copy(alpha = 0.88f),
            Color(0xFF051D12).copy(alpha = 0.88f)
        ),
        cardBorderColors = listOf(
            Color(0xFFFFE082),
            Color(0xFF81C784).copy(alpha = 0.60f),
            Color(0xFFFFD54F)
        ),
        nameTextColor = Color(0xFFFFDF7D),
        meaningTextColor = Color(0xFFE0F2E9),
        kalakiniTextColor = Color(0xFF81C784),
        quoteTextColor = Color(0xFFFFD54F),
        dividerColor = Color(0xFFD4AF37).copy(alpha = 0.35f),
        footerTextColor = Color(0xFFFFDF7D).copy(alpha = 0.90f),
        outerBorderColors = listOf(Color(0xFFFFDF7D), Color(0xFF1B5E20), Color(0xFFFFDF7D)),
        haloColors = listOf(
            Color(0xFFFFDF7D).copy(alpha = 0.70f),
            Color(0xFF81C784).copy(alpha = 0.35f),
            Color(0xFF072718).copy(alpha = 0.10f),
            Color.Transparent
        )
    ),
    MIDNIGHT(
        id = "midnight",
        title = "🪐 จักรวาลมงคล",
        bgFileName = "blessing_wallpaper_bg_midnight.jpg",
        fallbackResId = R.drawable.blessing_wallpaper_bg_midnight,
        cardBackground = Color(0xFF090E1D).copy(alpha = 0.88f),
        cardBackgroundColors = listOf(
            Color(0xFF090E1D).copy(alpha = 0.88f),
            Color(0xFF060914).copy(alpha = 0.88f)
        ),
        cardBorderColors = listOf(
            Color(0xFFFFDF7D),
            Color(0xFFD4AF37).copy(alpha = 0.65f),
            Color(0xFFFFDF7D)
        ),
        nameTextColor = Color(0xFFFFDF7D),
        meaningTextColor = Color(0xFFF0EBE1),
        kalakiniTextColor = Color(0xFF81C784),
        quoteTextColor = Color(0xFFFFD54F),
        dividerColor = Color(0xFFD4AF37).copy(alpha = 0.35f),
        footerTextColor = Color(0xFFFFDF7D).copy(alpha = 0.90f),
        outerBorderColors = listOf(Color(0xFFFFDF7D), Color(0xFF1B284F), Color(0xFFFFDF7D)),
        haloColors = listOf(
            Color(0xFFFFE599).copy(alpha = 0.85f),
            Color(0xFF6B8AFF).copy(alpha = 0.35f),
            Color(0xFF090E1D).copy(alpha = 0.10f),
            Color.Transparent
        )
    )
}
