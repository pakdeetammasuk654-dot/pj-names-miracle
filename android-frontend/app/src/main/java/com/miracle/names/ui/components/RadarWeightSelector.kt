package com.miracle.names.ui.components

import androidx.compose.foundation.Canvas
import androidx.compose.animation.core.animateFloat
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.size
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.PurpleHeader
import kotlin.math.abs
import kotlin.math.floor

@Composable
fun RadarWeightSelector(
    numerologyWeight: Float,
    semanticWeight: Float,
    soundWeight: Float,
    similarityWeight: Float,
    onWeightsChanged: (Float, Float, Float, Float) -> Unit,
    modifier: Modifier = Modifier,
    compact: Boolean = false,
    labelColor: Color = MaterialTheme.colorScheme.onSurface,
    showLabels: Boolean = true,
    showCornerIcons: Boolean = false
) {
    val selectorPulse = androidx.compose.animation.core.rememberInfiniteTransition(label = "selector_pulse")
    val pulseProgress = selectorPulse.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = androidx.compose.animation.core.infiniteRepeatable(
            animation = androidx.compose.animation.core.tween(
                durationMillis = 1050,
                easing = androidx.compose.animation.core.FastOutSlowInEasing
            ),
            repeatMode = androidx.compose.animation.core.RepeatMode.Reverse
        ),
        label = "selector_glow_pulse"
    )
    val graphBoxHeight = if (compact) 120.dp else 145.dp
    val labelFontSize = if (compact) 9.sp else 11.sp
    val cornerIconSize = if (compact) 13.sp else 15.sp

    Column(modifier = modifier) {
        if (!compact && showLabels) {
            Text(
                text = "⭐ เลขศาสตร์",
                modifier = Modifier.align(Alignment.CenterHorizontally),
                color = GoldPrimary,
                fontSize = 11.sp,
                fontWeight = FontWeight.Bold
            )
        }
        BoxWithConstraints(
            modifier = Modifier
                .fillMaxWidth()
                .height(graphBoxHeight)
        ) {
            val density = LocalDensity.current
            // Reserve a clear gutter around the plot when corner icons are shown.
            // The icons then sit outside the diamond instead of colliding with a vertex.
            val verticalInsetDp = when {
                showCornerIcons && compact -> 22.dp
                showCornerIcons -> 28.dp
                compact -> 14.dp
                else -> 18.dp
            }
            val horizontalInsetDp = when {
                showCornerIcons && compact -> 26.dp
                showCornerIcons -> 28.dp
                compact -> 22.dp
                else -> 28.dp
            }
            val verticalInsetPx = with(density) { verticalInsetDp.toPx() }
            val horizontalInsetPx = with(density) { horizontalInsetDp.toPx() }

            val currentOnWeightsChanged by rememberUpdatedState(onWeightsChanged)

            Canvas(
                modifier = Modifier
                    .fillMaxSize()
                    .pointerInput(horizontalInsetPx, verticalInsetPx) {
                        fun update(position: Offset) {
                            val weights = diamondWeightsAt(
                                position = position,
                                width = size.width.toFloat(),
                                height = size.height.toFloat(),
                                horizontalInset = horizontalInsetPx,
                                verticalInset = verticalInsetPx
                            )
                            currentOnWeightsChanged(weights[0], weights[1], weights[2], weights[3])
                        }
                        awaitEachGesture {
                            val down = awaitFirstDown(requireUnconsumed = false)
                            update(down.position)
                            down.consume()
                            while (true) {
                                val event = awaitPointerEvent()
                                val change = event.changes.firstOrNull { it.id == down.id } ?: break
                                if (!change.pressed) break
                                update(change.position)
                                change.consume()
                            }
                        }
                    }
            ) {
                val cx = size.width / 2f
                val cy = size.height / 2f
                val top = Offset(cx, verticalInsetPx)
                val bottom = Offset(cx, size.height - verticalInsetPx)
                val left = Offset(horizontalInsetPx, cy)
                val right = Offset(size.width - horizontalInsetPx, cy)

                val diamond = Path().apply {
                    moveTo(top.x, top.y)
                    lineTo(right.x, right.y)
                    lineTo(bottom.x, bottom.y)
                    lineTo(left.x, left.y)
                    close()
                }

                drawPath(
                    path = diamond,
                    brush = Brush.verticalGradient(
                        listOf(
                            Color(0xFFFFE58A).copy(alpha = 0.34f),
                            Color(0xFF382014).copy(alpha = 0.50f)
                        )
                    )
                )
                drawPath(
                    path = diamond,
                    brush = Brush.linearGradient(
                        listOf(
                            Color(0xFFFFE58A),
                            Color(0xFFF2B94B),
                            Color(0xFF4DB6AC),
                            Color(0xFF7DD3A8),
                            Color(0xFFFFE58A)
                        )
                    ),
                    style = Stroke(width = 2.5f)
                )

                // Draw cross lines for the axes
                drawLine(
                    color = Color.White.copy(alpha = 0.22f),
                    start = top,
                    end = bottom,
                    strokeWidth = 1.5f
                )
                drawLine(
                    color = Color.White.copy(alpha = 0.22f),
                    start = left,
                    end = right,
                    strokeWidth = 1.5f
                )

                val selector = Offset(
                    (top.x * numerologyWeight) + (bottom.x * soundWeight) + (left.x * semanticWeight) + (right.x * similarityWeight),
                    (top.y * numerologyWeight) + (bottom.y * soundWeight) + (left.y * semanticWeight) + (right.y * similarityWeight)
                )
                val pulse = pulseProgress.value
                val baseHaloRadius = 17.dp.toPx()
                val animatedHaloRadius = baseHaloRadius + (5.dp.toPx() * pulse)
                val outerRingRadius = 12.dp.toPx() + (4.dp.toPx() * pulse)
                drawCircle(
                    brush = Brush.radialGradient(
                        colors = listOf(
                            Color(0xFFFFF8D5).copy(alpha = 0.92f - (pulse * 0.12f)),
                            Color(0xFFFFC84A).copy(alpha = 0.58f - (pulse * 0.18f)),
                            Color.Transparent
                        ),
                        center = selector,
                        radius = animatedHaloRadius
                    ),
                    radius = animatedHaloRadius,
                    center = selector
                )
                drawCircle(
                    color = Color(0xFFFFD65A).copy(alpha = 0.48f * (1f - pulse)),
                    radius = outerRingRadius,
                    center = selector,
                    style = Stroke(width = 2.dp.toPx())
                )
                drawCircle(Color.White, radius = 7.5.dp.toPx(), center = selector)
                drawCircle(Color(0xFFF59E0B), radius = 4.dp.toPx(), center = selector)
                drawCircle(Color(0xFFFFFDF0), radius = 1.8.dp.toPx(), center = selector)
            }

            if (showLabels) {
                if (compact) {
                    Text(
                        text = "⭐ เลข",
                        modifier = Modifier.align(Alignment.TopCenter),
                        fontSize = labelFontSize,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFFFDE047)
                    )
                }
                Text(
                    text = "🎯 ความหมาย",
                    modifier = Modifier.align(Alignment.CenterStart),
                    fontSize = labelFontSize,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFFF2B94B)
                )
                Text(
                    text = "🔍 เอกลักษณ์",
                    modifier = Modifier.align(Alignment.CenterEnd),
                    fontSize = labelFontSize,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF4DB6AC)
                )
                Text(
                    text = "🎶 ไพเราะ",
                    modifier = Modifier.align(Alignment.BottomCenter),
                    fontSize = labelFontSize,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF7DD3A8)
                )
            }
            if (showCornerIcons) {
                // Top icon
                Text(
                    text = "⭐",
                    fontSize = cornerIconSize,
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .offset(y = 0.dp)
                )
                // Left icon
                Text(
                    text = "🎯",
                    fontSize = cornerIconSize,
                    modifier = Modifier
                        .align(Alignment.CenterStart)
                        .offset(x = 0.dp)
                )
                // Right icon
                Text(
                    text = "🔍",
                    fontSize = cornerIconSize,
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .offset(x = 0.dp)
                )
                // Bottom icon
                Text(
                    text = "🎶",
                    fontSize = cornerIconSize,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .offset(y = 0.dp)
                )
            }
        }
    }
}

internal fun diamondWeightsAt(
    position: Offset,
    width: Float,
    height: Float,
    horizontalInset: Float = 28f,
    verticalInset: Float = 18f
): List<Float> {
    val cx = width / 2f
    val cy = height / 2f
    val rx = (cx - horizontalInset).coerceAtLeast(1f)
    val ry = (cy - verticalInset).coerceAtLeast(1f)

    var nx = (position.x - cx) / rx
    var ny = (position.y - cy) / ry

    val scale = abs(nx) + abs(ny)
    if (scale > 1f) {
        nx /= scale
        ny /= scale
    }

    val u = (nx + ny + 1f) / 2f
    val v = (nx - ny + 1f) / 2f

                val sound = u * (1f - v)
                val numerology = (1f - u) * v
                val semantic = (1f - u) * (1f - v)
                val similarity = u * v

    return quantizeWeights(numerology, semantic, sound, similarity)
}

internal fun diamondWeightsAtRadius(
    position: Offset,
    width: Float,
    height: Float,
    radius: Float
): List<Float> {
    return diamondWeightsAt(
        position = position,
        width = width,
        height = height,
        horizontalInset = (width / 2f) - radius,
        verticalInset = (height / 2f) - radius
    )
}

internal fun diamondRadius(width: Float, height: Float, withCornerIcons: Boolean = false): Float {
    val horizontalInset = 28f
    val verticalInset = if (withCornerIcons) 28f else 18f
    return minOf(
        (width / 2f) - horizontalInset,
        (height / 2f) - verticalInset
    ).coerceAtLeast(0f)
}

private fun quantizeWeights(numerology: Float, semantic: Float, sound: Float, similarity: Float): List<Float> {
    val rawUnits = listOf(numerology, semantic, sound, similarity).map { it * 20f }
    val units = rawUnits.map { floor(it).toInt() }.toMutableList()
    var remaining = 20 - units.sum()
    rawUnits
        .mapIndexed { index, value -> index to (value - floor(value)) }
        .sortedByDescending { it.second }
        .forEach { (index, _) ->
            if (remaining > 0) {
                units[index] += 1
                remaining -= 1
            }
        }
    return listOf(units[0] / 20f, units[1] / 20f, units[2] / 20f, units[3] / 20f)
}
