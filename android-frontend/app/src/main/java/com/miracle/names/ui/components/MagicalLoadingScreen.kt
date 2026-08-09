package com.miracle.names.ui.components

import androidx.compose.animation.Crossfade
import androidx.compose.animation.core.*
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.draw.scale
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.GoldLight
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.PurpleAccent
import com.miracle.names.ui.theme.PurpleDark
import com.miracle.names.ui.theme.PurpleHeader
import kotlinx.coroutines.delay

@Composable
fun MagicalLoadingScreen() {
    val isDark = isSystemInDarkTheme()

    val messages = remember {
        listOf(
            "✨ กำลังเปิดม่านรายชื่อมงคล...",
            "🔮 กำลังร้อยเรียงพลังของตัวเลข...",
            "🌟 กำลังค้นหาชื่อที่เหมาะกับคุณ...",
            "📜 คำนวณเลขศาสตร์และพลังเงา..."
        )
    }

    var messageIndex by remember { mutableIntStateOf(0) }

    LaunchedEffect(Unit) {
        while (true) {
            delay(1400)
            messageIndex = (messageIndex + 1) % messages.size
        }
    }

    val infiniteTransition = rememberInfiniteTransition(label = "magical_loading_screen")

    // Smooth infinite rotation for main star
    val rotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(4000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "rotation"
    )

    // Pulsing scale for star accent
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 0.88f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulse"
    )

    // Shimmer offset for sweeping multi-color gradient
    val shimmerOffset by infiniteTransition.animateFloat(
        initialValue = -600f,
        targetValue = 1400f,
        animationSpec = infiniteRepeatable(
            animation = tween(1500, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmer"
    )

    // Vibrant Iridescent Magical Shimmer Colors
    val baseSkeletonColor = if (isDark) Color(0xFF2E1C3B) else Color(0xFFEDE7F3)
    val shimmerColors = listOf(
        baseSkeletonColor,
        Color(0xFFE9D5FF).copy(alpha = 0.65f), // Soft Lavender Glow
        Color(0xFFFDE68A).copy(alpha = 0.85f), // Radiant Gold
        Color(0xFFF472B6).copy(alpha = 0.75f), // Vibrant Magenta/Pink
        Color(0xFF818CF8).copy(alpha = 0.65f), // Celestial Indigo
        baseSkeletonColor
    )

    val shimmerBrush = Brush.linearGradient(
        colors = shimmerColors,
        start = Offset(shimmerOffset, 0f),
        end = Offset(shimmerOffset + 450f, 250f)
    )

    val progressBrush = Brush.horizontalGradient(
        colors = listOf(
            GoldPrimary,
            Color(0xFFEC4899),
            Color(0xFF8B5CF6),
            Color(0xFF06B6D4),
            GoldPrimary
        ),
        startX = shimmerOffset,
        endX = shimmerOffset + 600f
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        // Floating Sparkle Particles in Background
        MagicalParticleBackground(infiniteTransition = infiniteTransition)

        // Radial Background Aura Glow
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(300.dp)
                .background(
                    Brush.radialGradient(
                        colors = listOf(
                            PurpleAccent.copy(alpha = if (isDark) 0.25f else 0.12f),
                            GoldPrimary.copy(alpha = if (isDark) 0.15f else 0.08f),
                            Color.Transparent
                        ),
                        center = Offset(500f, 100f),
                        radius = 800f
                    )
                )
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 16.dp)
        ) {
            Spacer(modifier = Modifier.height(20.dp))

            // Magical Status Header Banner
            MagicalStatusBanner(
                message = messages[messageIndex],
                rotation = rotation,
                pulseScale = pulseScale,
                progressBrush = progressBrush
            )

            Spacer(modifier = Modifier.height(18.dp))

            // Table Header Skeleton Bar
            MagicalTableHeaderSkeleton(shimmerBrush = shimmerBrush, isDark = isDark)

            Spacer(modifier = Modifier.height(10.dp))

            // Result Cards Skeleton List
            Column(
                verticalArrangement = Arrangement.spacedBy(10.dp)
            ) {
                repeat(5) { index ->
                    MagicalSkeletonNameCard(
                        brush = shimmerBrush,
                        isFirst = index == 0,
                        rank = index + 1,
                        isDark = isDark
                    )
                }
            }
        }
    }
}

@Composable
private fun MagicalStatusBanner(
    message: String,
    rotation: Float,
    pulseScale: Float,
    progressBrush: Brush
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(18.dp),
        color = PurpleHeader,
        shadowElevation = 6.dp,
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
        )
    ) {
        Column(
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
                .padding(top = 16.dp, bottom = 12.dp, start = 16.dp, end = 16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.Center,
                modifier = Modifier.fillMaxWidth()
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Icon(
                        imageVector = Icons.Default.AutoAwesome,
                        contentDescription = null,
                        tint = GoldPrimary,
                        modifier = Modifier
                            .size(30.dp)
                            .scale(pulseScale)
                            .rotate(rotation)
                    )
                    Icon(
                        imageVector = Icons.Default.Star,
                        contentDescription = null,
                        tint = Color(0xFFF472B6),
                        modifier = Modifier
                            .size(14.dp)
                            .rotate(-rotation * 1.5f)
                    )
                }

                Spacer(modifier = Modifier.width(12.dp))

                Crossfade(
                    targetState = message,
                    animationSpec = tween(400),
                    label = "message_crossfade"
                ) { currentMessage ->
                    Text(
                        text = currentMessage,
                        color = Color.White,
                        fontSize = 15.sp,
                        fontWeight = FontWeight.Bold,
                        textAlign = TextAlign.Start
                    )
                }
            }

            Spacer(modifier = Modifier.height(14.dp))

            // Animated Iridescent Progress Line
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(4.dp)
                    .clip(RoundedCornerShape(2.dp))
                    .background(Color.White.copy(alpha = 0.15f))
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(progressBrush)
                )
            }
        }
    }
}

@Composable
private fun MagicalTableHeaderSkeleton(shimmerBrush: Brush, isDark: Boolean) {
    val headerBg = if (isDark) Color(0xFF271735) else Color(0xFFEFE8F5)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(34.dp)
            .clip(RoundedCornerShape(8.dp))
            .background(headerBg)
            .padding(horizontal = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Box(
            modifier = Modifier
                .width(28.dp)
                .height(12.dp)
                .clip(RoundedCornerShape(4.dp))
                .background(shimmerBrush)
        )
        Box(
            modifier = Modifier
                .width(60.dp)
                .height(12.dp)
                .clip(RoundedCornerShape(4.dp))
                .background(shimmerBrush)
        )
        Box(
            modifier = Modifier
                .width(80.dp)
                .height(12.dp)
                .clip(RoundedCornerShape(4.dp))
                .background(shimmerBrush)
        )
        Box(
            modifier = Modifier
                .width(40.dp)
                .height(12.dp)
                .clip(RoundedCornerShape(4.dp))
                .background(shimmerBrush)
        )
    }
}

@Composable
private fun MagicalSkeletonNameCard(
    brush: Brush,
    isFirst: Boolean,
    rank: Int,
    isDark: Boolean
) {
    val shape = RoundedCornerShape(16.dp)
    val cardBg = if (isFirst) {
        if (isDark) Color(0xFF3B234A) else Color(0xFFFFFBEB)
    } else {
        if (isDark) Color(0xFF281A34) else Color(0xFFFFFFFF)
    }

    val borderModifier = if (isFirst) {
        Modifier.border(
            BorderStroke(
                1.5.dp,
                Brush.horizontalGradient(
                    listOf(
                        GoldPrimary,
                        Color(0xFFF472B6),
                        PurpleAccent
                    )
                )
            ),
            shape
        )
    } else {
        Modifier.border(
            BorderStroke(
                0.8.dp,
                if (isDark) Color(0xFF442D56) else Color(0xFFE8DFEE)
            ),
            shape
        )
    }

    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .then(borderModifier),
        shape = shape,
        color = cardBg,
        shadowElevation = if (isFirst) 4.dp else 1.dp
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Rank Number Skeleton
            Box(
                modifier = Modifier.width(36.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "#$rank",
                    color = if (isFirst) GoldPrimary else (if (isDark) GoldLight.copy(alpha = 0.5f) else Color.Gray),
                    fontWeight = FontWeight.Bold,
                    fontSize = 14.sp
                )
            }

            Spacer(modifier = Modifier.width(8.dp))

            // Main Name Info & PairCircles Skeleton
            Column(modifier = Modifier.weight(1f)) {
                // Name Title Bar Skeleton
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth(if (isFirst) 0.55f else 0.45f)
                            .height(18.dp)
                            .clip(RoundedCornerShape(6.dp))
                            .background(brush)
                    )

                    if (isFirst) {
                        Spacer(modifier = Modifier.width(8.dp))
                        // "Top Match" Badge Placeholder
                        Box(
                            modifier = Modifier
                                .width(56.dp)
                                .height(16.dp)
                                .clip(RoundedCornerShape(8.dp))
                                .background(
                                    Brush.horizontalGradient(
                                        listOf(GoldPrimary.copy(alpha = 0.4f), Color(0xFFF472B6).copy(alpha = 0.4f))
                                    )
                                )
                        )
                    }
                }

                Spacer(modifier = Modifier.height(8.dp))

                // Pair Circles Skeleton (Product Rule: colored pill representation for number pairs)
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    repeat(3) {
                        Box(
                            modifier = Modifier
                                .width(36.dp)
                                .height(18.dp)
                                .clip(RoundedCornerShape(9.dp))
                                .background(brush)
                        )
                    }
                }

                Spacer(modifier = Modifier.height(6.dp))

                // Numerology / Shadow Power Summary Line Skeleton
                Box(
                    modifier = Modifier
                        .fillMaxWidth(0.85f)
                        .height(11.dp)
                        .clip(RoundedCornerShape(4.dp))
                        .background(brush)
                )
            }

            Spacer(modifier = Modifier.width(10.dp))

            // Score Pill Badge Skeleton
            Box(
                modifier = Modifier
                    .width(44.dp)
                    .height(26.dp)
                    .clip(RoundedCornerShape(13.dp))
                    .background(
                        if (isFirst) {
                            Brush.linearGradient(
                                listOf(GoldPrimary.copy(alpha = 0.35f), Color(0xFFF472B6).copy(alpha = 0.35f))
                            )
                        } else {
                            brush
                        }
                    )
            )
        }
    }
}

@Composable
private fun MagicalParticleBackground(infiniteTransition: InfiniteTransition) {
    val alpha1 by infiniteTransition.animateFloat(
        initialValue = 0.2f,
        targetValue = 0.85f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "particle_1"
    )
    val alpha2 by infiniteTransition.animateFloat(
        initialValue = 0.7f,
        targetValue = 0.15f,
        animationSpec = infiniteRepeatable(
            animation = tween(1700, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "particle_2"
    )

    Canvas(modifier = Modifier.fillMaxSize()) {
        // Floating Sparkle 1 (Top Left)
        drawCircle(
            color = Color(0xFFFDE68A).copy(alpha = alpha1),
            radius = 6.dp.toPx(),
            center = Offset(size.width * 0.12f, size.height * 0.08f)
        )
        // Floating Sparkle 2 (Top Right)
        drawCircle(
            color = Color(0xFFF472B6).copy(alpha = alpha2),
            radius = 8.dp.toPx(),
            center = Offset(size.width * 0.88f, size.height * 0.14f)
        )
        // Floating Sparkle 3 (Mid Right)
        drawCircle(
            color = Color(0xFFA78BFA).copy(alpha = alpha1),
            radius = 5.dp.toPx(),
            center = Offset(size.width * 0.92f, size.height * 0.45f)
        )
        // Floating Sparkle 4 (Lower Left)
        drawCircle(
            color = Color(0xFF38BDF8).copy(alpha = alpha2),
            radius = 7.dp.toPx(),
            center = Offset(size.width * 0.08f, size.height * 0.62f)
        )
    }
}
