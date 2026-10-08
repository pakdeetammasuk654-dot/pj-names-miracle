package com.miracle.names.ui.auspicious

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.drawscope.rotate
import androidx.compose.ui.graphics.nativeCanvas
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.TaksaCalculationResponse
import com.miracle.names.data.model.TaksaRole
import com.miracle.names.ui.theme.GoldLight
import com.miracle.names.ui.theme.GoldPrimary
import kotlin.math.cos
import kotlin.math.atan2
import kotlin.math.min
import kotlin.math.sin
import kotlin.math.sqrt

private val wheelMutedText = Color(0xFFD6C8BE)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaksaWheelPanel(
    result: TaksaCalculationResponse,
    modifier: Modifier = Modifier
) {
    var selectedRole by remember { mutableStateOf<TaksaRole?>(null) }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 4.dp, vertical = 6.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Text(
            "กระดานทดสำหรับวิเคราะห์ฤกษ์ยาม",
            color = wheelMutedText,
            fontSize = 11.sp,
            lineHeight = 17.sp
        )
        Row(verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f)) {
                Text(
                    if (result.isTodayDefault) "✨ ทักษาพลังประจำวัน" else "🧭 ทักษาจรประจำปีของคุณ",
                    color = GoldLight,
                    fontSize = 19.sp,
                    fontWeight = FontWeight.Bold
                )
                val summary = when {
                    result.isTodayDefault -> "แสดงพลังของวันนี้ · ระบุวันเกิดเพื่อดูผลเฉพาะบุคคล"
                    result.isCenterCell -> "เกิด${result.birthDayName} · อายุย่าง ${result.progressedAge} ปี · ตกตากลาง ๙"
                    else -> "เกิด${result.birthDayName} · อายุย่าง ${result.progressedAge} ปี · ตก${result.landedPlanetName}"
                }
                Text(summary, color = wheelMutedText, fontSize = 12.sp, lineHeight = 18.sp)
            }
        }

        Text(
            "แตะแต่ละช่องเพื่อดูความหมายต่อฤกษ์ของคุณ",
            color = GoldPrimary,
            fontSize = 11.sp,
            fontWeight = FontWeight.SemiBold
        )
        TaksaWheel(
            result = result,
            spinRevision = 0,
            onRoleSelected = { selectedRole = it }
        )
    }

    selectedRole?.let { role ->
        TaksaRoleDetailSheet(role = role, onDismiss = { selectedRole = null })
    }
}

@Composable
private fun TaksaWheel(
    result: TaksaCalculationResponse,
    spinRevision: Int,
    onRoleSelected: (TaksaRole) -> Unit
) {
    val animatedRotation = remember { Animatable(result.wheelRotationDegrees) }
    LaunchedEffect(result, spinRevision) {
        val target = result.wheelRotationDegrees
        animatedRotation.snapTo(target - 1080f)
        animatedRotation.animateTo(
            targetValue = target,
            animationSpec = tween(durationMillis = 1800, easing = FastOutSlowInEasing)
        )
    }
    val density = LocalDensity.current
    val planetTextSize = with(density) { 13.sp.toPx() }
    val roleTextSize = with(density) { 10.sp.toPx() }
    val roles = result.allRoles

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .aspectRatio(1f)
            .pointerInput(roles) {
                detectTapGestures { tap ->
                    val centerX = size.width / 2f
                    val centerY = size.height / 2f
                    val dx = tap.x - centerX
                    val dy = tap.y - centerY
                    val outerRadius = min(size.width, size.height) / 2f - 8.dp.toPx()
                    val innerRadius = outerRadius * .39f
                    val distance = sqrt(dx * dx + dy * dy)
                    if (distance in innerRadius..outerRadius && roles.isNotEmpty()) {
                        val degrees = Math.toDegrees(atan2(dy.toDouble(), dx.toDouble()))
                        val normalized = ((degrees + 112.5) % 360.0 + 360.0) % 360.0
                        val index = (normalized / 45.0).toInt().coerceIn(roles.indices)
                        onRoleSelected(roles[index])
                    }
                }
            }
            .semantics {
                contentDescription = if (result.isTodayDefault) {
                    "วงล้อทักษาพลังประจำวัน ตก${result.landedPlanetName}"
                } else {
                    "วงล้อทักษาจร อายุย่าง ${result.progressedAge} ปี ตก${result.landedPlanetName}"
                }
            },
        contentAlignment = Alignment.Center
    ) {
        Canvas(Modifier.fillMaxWidth().aspectRatio(1f).padding(4.dp)) {
            val diameter = min(size.width, size.height)
            val center = Offset(size.width / 2f, size.height / 2f)
            val outerRadius = diameter / 2f - 8.dp.toPx()
            val innerRadius = outerRadius * .39f
            val relativeRotation = animatedRotation.value - result.wheelRotationDegrees

            drawCircle(Color(0xFF140B07), outerRadius + 5.dp.toPx(), center)
            drawCircle(GoldPrimary, outerRadius + 3.dp.toPx(), center, style = androidx.compose.ui.graphics.drawscope.Stroke(2.dp.toPx()))

            rotate(relativeRotation, center) {
                roles.forEachIndexed { index, role ->
                    val color = roleSectorColor(role, index)
                    drawArc(
                        color = color,
                        startAngle = -112.5f + index * 45f,
                        sweepAngle = 44.3f,
                        useCenter = true,
                        topLeft = Offset(center.x - outerRadius, center.y - outerRadius),
                        size = Size(outerRadius * 2, outerRadius * 2)
                    )
                }
                drawCircle(Color(0xFF140B07), innerRadius, center)
                drawCircle(GoldPrimary.copy(alpha = .75f), innerRadius, center, style = androidx.compose.ui.graphics.drawscope.Stroke(1.5.dp.toPx()))
            }

            roles.forEachIndexed { index, role ->
                val angle = Math.toRadians((-90.0 + index * 45.0))
                val planetRadius = outerRadius * .76f
                val roleRadius = outerRadius * .53f
                val planetPoint = Offset(
                    center.x + (cos(angle) * planetRadius).toFloat(),
                    center.y + (sin(angle) * planetRadius).toFloat()
                )
                val rolePoint = Offset(
                    center.x + (cos(angle) * roleRadius).toFloat(),
                    center.y + (sin(angle) * roleRadius).toFloat()
                )
                drawContext.canvas.nativeCanvas.apply {
                    val planetPaint = android.graphics.Paint(android.graphics.Paint.ANTI_ALIAS_FLAG).apply {
                        color = android.graphics.Color.WHITE
                        textAlign = android.graphics.Paint.Align.CENTER
                        textSize = planetTextSize
                        typeface = android.graphics.Typeface.create(android.graphics.Typeface.DEFAULT, android.graphics.Typeface.BOLD)
                    }
                    val rolePaint = android.graphics.Paint(android.graphics.Paint.ANTI_ALIAS_FLAG).apply {
                        color = roleTextColor(role).toArgb()
                        textAlign = android.graphics.Paint.Align.CENTER
                        textSize = roleTextSize
                        typeface = android.graphics.Typeface.create(android.graphics.Typeface.DEFAULT, android.graphics.Typeface.BOLD)
                    }
                    drawText(
                        "${role.symbol} ${role.dayName.substringBefore(" (")}",
                        planetPoint.x,
                        planetPoint.y + planetTextSize * .35f,
                        planetPaint
                    )
                    drawText(role.roleName.removeSuffix("จร"), rolePoint.x, rolePoint.y + roleTextSize * .35f, rolePaint)
                }
            }

            val pointer = Path().apply {
                moveTo(center.x, 2.dp.toPx())
                lineTo(center.x - 8.dp.toPx(), 18.dp.toPx())
                lineTo(center.x + 8.dp.toPx(), 18.dp.toPx())
                close()
            }
            drawPath(pointer, GoldPrimary)
        }

        Column(
            modifier = Modifier
                .size(112.dp)
                .background(Color(0xFF26160E), CircleShape)
                .padding(horizontal = 8.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(if (result.isTodayDefault) "พลังทักษา" else "อายุย่าง", color = GoldPrimary, fontSize = 11.sp, fontWeight = FontWeight.Bold)
            Text(
                if (result.isTodayDefault) "วันนี้" else "${result.progressedAge} ปี",
                color = Color.White,
                fontSize = 18.sp,
                fontWeight = FontWeight.ExtraBold
            )
            Text(
                if (result.isCenterCell) "ตากลาง ๙" else "${result.landedPlanetSymbol} ดาว ${result.landedPlanetNum}",
                color = Color(0xFFFFDF85),
                fontSize = 10.sp,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun TaksaRoleDetailSheet(role: TaksaRole, onDismiss: () -> Unit) {
    val accent = roleTextColor(role)
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = Color(0xFF160C08)
    ) {
        Column(
            modifier = Modifier.fillMaxWidth().padding(start = 22.dp, end = 22.dp, bottom = 32.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(role.symbol, fontSize = 34.sp)
                Column(Modifier.padding(start = 12.dp)) {
                    Text(role.roleName, color = accent, fontSize = 22.sp, fontWeight = FontWeight.Bold)
                    Text(role.dayName, color = Color.White, fontSize = 14.sp)
                }
            }
            Text(
                if (role.impact == "caution") "ควรระวังในการเลือกฤกษ์" else "พลังสนับสนุนการเลือกฤกษ์",
                color = if (role.impact == "caution") Color(0xFFFFA6A6) else Color(0xFFFFDF85),
                fontWeight = FontWeight.Bold
            )
            Text("ความหมาย", color = GoldPrimary, fontWeight = FontWeight.Bold)
            Text(role.meaning.orEmpty().ifBlank { role.description }, color = Color(0xFFE7E2CF), lineHeight = 22.sp)
            Text("ความสัมพันธ์กับฤกษ์ของคุณ", color = GoldPrimary, fontWeight = FontWeight.Bold)
            Text(
                role.auspiciousGuidance.orEmpty().ifBlank { "${role.dayName} สัมพันธ์กับ${role.roleName}ในการวิเคราะห์ฤกษ์ของคุณ" },
                color = Color(0xFFE7E2CF),
                lineHeight = 22.sp
            )
        }
    }
}

private fun roleSectorColor(role: TaksaRole, index: Int): Color = when {
    role.roleName.contains("ศรี") -> Color(0xFF4A4914)
    role.roleName.contains("เดช") -> Color(0xFF563517)
    role.roleName.contains("มนตรี") -> Color(0xFF3D2517)
    role.roleName.contains("กาล") -> Color(0xFF541B1B)
    index % 2 == 0 -> Color(0xFF382014)
    else -> Color(0xFF26160E)
}

private fun roleTextColor(role: TaksaRole): Color = when {
    role.roleName.contains("ศรี") -> Color(0xFFF6E56B)
    role.roleName.contains("เดช") -> Color(0xFFFFC987)
    role.roleName.contains("มนตรี") -> Color(0xFFFFE0A0)
    role.roleName.contains("กาล") -> Color(0xFFFFA6A6)
    else -> GoldPrimary
}

private fun Color.toArgb(): Int = android.graphics.Color.argb(
    (alpha * 255).toInt(),
    (red * 255).toInt(),
    (green * 255).toInt(),
    (blue * 255).toInt()
)
