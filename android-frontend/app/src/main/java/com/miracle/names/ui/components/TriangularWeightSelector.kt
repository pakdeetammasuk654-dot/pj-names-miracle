package com.miracle.names.ui.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.PurpleHeader
import kotlin.math.floor

@Composable
fun TriangularWeightSelector(
    numerologyWeight: Float,
    semanticWeight: Float,
    soundWeight: Float,
    onWeightsChanged: (Float, Float, Float) -> Unit,
	modifier: Modifier = Modifier,
	compact: Boolean = false,
	labelColor: Color = MaterialTheme.colorScheme.onSurface,
	showLabels: Boolean = true,
	showCornerIcons: Boolean = false
) {
	val graphHeight = if (compact) 82.dp else 150.dp
	val graphBoxHeight = if (showCornerIcons) graphHeight + 34.dp else graphHeight
	val labelFontSize = if (compact) 9.sp else 11.sp
	val cornerIconSize = if (compact) 16.sp else 18.sp
	val graphGold = Color(0xFFD09A00)
	val iconShadow = Color(0xFF6B4A00).copy(alpha = 0.62f)
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
        Box(
			modifier = Modifier
				.fillMaxWidth()
				.height(graphBoxHeight)
		) {
			Canvas(
				modifier = Modifier
					.fillMaxSize()
					.pointerInput(Unit) {
						detectTapGestures { position ->
							val weights = triangleWeightsAt(
								position,
								size.width.toFloat(),
								size.height.toFloat(),
								showCornerIcons
							)
							onWeightsChanged(weights.first, weights.second, weights.third)
						}
					}
					.pointerInput(Unit) {
						fun update(position: Offset) {
							val weights = triangleWeightsAt(
								position,
								size.width.toFloat(),
								size.height.toFloat(),
								showCornerIcons
							)
							onWeightsChanged(weights.first, weights.second, weights.third)
						}
                        detectDragGestures(
                            onDragStart = { update(it) },
                            onDrag = { change, _ ->
                                update(change.position)
                                change.consume()
                            }
                        )
                    }
            ) {
				val iconPadding = if (showCornerIcons) 38f else 0f
				val top = Offset(size.width / 2f, 8f + iconPadding)
				val left = Offset(28f + iconPadding, size.height - 30f)
				val right = Offset(size.width - 28f - iconPadding, size.height - 30f)
                val triangle = Path().apply {
                    moveTo(top.x, top.y)
                    lineTo(left.x, left.y)
                    lineTo(right.x, right.y)
                    close()
                }

				drawPath(
					path = triangle,
					brush = Brush.verticalGradient(
						listOf(
							Color(0xFFFDE047).copy(alpha = 0.40f),
							Color(0xFFF472B6).copy(alpha = 0.30f),
							Color(0xFF38BDF8).copy(alpha = 0.35f)
						)
					)
				)
				drawPath(
					path = triangle,
					brush = Brush.linearGradient(
						listOf(
							Color(0xFFFDE047),
							Color(0xFFF472B6),
							Color(0xFF38BDF8),
							Color(0xFFFDE047)
						)
					),
					style = Stroke(width = 3.6f)
				)
                for (fraction in listOf(1f / 3f, 2f / 3f)) {
					drawLine(
						brush = Brush.horizontalGradient(
							listOf(
								Color(0xFFF472B6).copy(alpha = 0.45f),
								Color(0xFF38BDF8).copy(alpha = 0.45f)
							)
						),
						start = top + ((left - top) * fraction),
						end = top + ((right - top) * fraction),
						strokeWidth = 1.8f
					)
				}

                val selector = Offset(
                    (top.x * numerologyWeight) + (left.x * semanticWeight) + (right.x * soundWeight),
                    (top.y * numerologyWeight) + (left.y * semanticWeight) + (right.y * soundWeight)
                )
				drawCircle(
					brush = Brush.radialGradient(
						colors = listOf(
							Color(0xFFFDE047).copy(alpha = 0.85f),
							Color(0xFFF472B6).copy(alpha = 0.50f),
							Color.Transparent
						),
						center = selector,
						radius = if (compact) 26f else 32f
					),
					radius = if (compact) 26f else 32f,
					center = selector
				)
				drawCircle(Color.White, radius = 11f, center = selector)
				drawCircle(Color(0xFFD97706), radius = 7.5f, center = selector)
				drawCircle(Color(0xFFFFFDF0), radius = 3.5f, center = selector)
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
                    modifier = Modifier.align(Alignment.BottomStart),
                    fontSize = labelFontSize,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFFF472B6)
                )
                Text(
                    text = "🎶 ไพเราะ",
                    modifier = Modifier.align(Alignment.BottomEnd),
                    fontSize = labelFontSize,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF38BDF8)
				)
			}
			if (showCornerIcons) {
				CornerIcon(
					text = "⭐",
					modifier = Modifier
						.align(Alignment.TopCenter)
						.offset(y = (-8).dp),
					fontSize = cornerIconSize,
					color = Color(0xFFFDE047),
					shadowColor = iconShadow
				)
				CornerIcon(
					text = "🎯",
					modifier = Modifier
						.align(Alignment.BottomStart)
						.offset(x = (-2).dp, y = 2.dp),
					fontSize = cornerIconSize,
					color = Color(0xFFF472B6),
					shadowColor = iconShadow
				)
				CornerIcon(
					text = "🎶",
					modifier = Modifier
						.align(Alignment.BottomEnd)
						.offset(x = 2.dp, y = 2.dp),
					fontSize = cornerIconSize,
					color = Color(0xFF38BDF8),
					shadowColor = iconShadow
				)
			}
		}
	}
}

@Composable
private fun CornerIcon(
	text: String,
	modifier: Modifier,
	fontSize: androidx.compose.ui.unit.TextUnit,
	color: Color,
	shadowColor: Color
) {
	Box(modifier = modifier) {
		Text(
			text = text,
			modifier = Modifier.offset(x = 1.dp, y = 1.dp),
			fontSize = fontSize,
			fontWeight = FontWeight.Bold,
			color = shadowColor
		)
		Text(
			text = text,
			fontSize = fontSize,
			fontWeight = FontWeight.Bold,
			color = color
		)
	}
}

private fun triangleWeightsAt(
	position: Offset,
	width: Float,
	height: Float,
	withCornerIcons: Boolean = false
): Triple<Float, Float, Float> {
	val iconPadding = if (withCornerIcons) 38f else 0f
	val top = Offset(width / 2f, 8f + iconPadding)
	val left = Offset(28f + iconPadding, height - 30f)
	val right = Offset(width - 28f - iconPadding, height - 30f)
    val denominator = ((left.y - right.y) * (top.x - right.x)) +
        ((right.x - left.x) * (top.y - right.y))
    if (denominator == 0f) return Triple(0.45f, 0.35f, 0.20f)

    var numerology = (((left.y - right.y) * (position.x - right.x)) +
        ((right.x - left.x) * (position.y - right.y))) / denominator
    var meaning = (((right.y - top.y) * (position.x - right.x)) +
        ((top.x - right.x) * (position.y - right.y))) / denominator
    var sound = 1f - numerology - meaning
    numerology = numerology.coerceAtLeast(0f)
    meaning = meaning.coerceAtLeast(0f)
    sound = sound.coerceAtLeast(0f)
    val total = numerology + meaning + sound
    if (total <= 0f) return Triple(0.45f, 0.35f, 0.20f)

    return quantizeWeights(numerology / total, meaning / total, sound / total)
}

private fun quantizeWeights(numerology: Float, meaning: Float, sound: Float): Triple<Float, Float, Float> {
    val rawUnits = listOf(numerology, meaning, sound).map { it * 20f }
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
    return Triple(units[0] / 20f, units[1] / 20f, units[2] / 20f)
}
