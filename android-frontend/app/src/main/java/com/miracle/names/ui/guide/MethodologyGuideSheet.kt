package com.miracle.names.ui.guide

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.PurpleHeader

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MethodologyGuideSheet(
    onDismiss: () -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
        containerColor = PurpleHeader
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = "📖 MIRACLE NAMES METHODOLOGY GUIDE",
                color = GoldPrimary,
                fontWeight = FontWeight.Bold,
                fontSize = 16.sp
            )
            Text(
                text = "คู่มือกระบวนการคิดชื่อมงคล & หลักเกณฑ์จัดอันดับ",
                color = Color.White,
                fontSize = 13.sp
            )

            Divider(color = GoldPrimary.copy(alpha = 0.5f))

            Text(
                text = "1. หลักการกาลกิณี (Kalakini Safety - 100% Safe)",
                color = GoldPrimary,
                fontWeight = FontWeight.Bold,
                fontSize = 14.sp
            )
            Text(
                text = "• ระบบกรองอักขระกาลกิณีออก 100% เสมอตามวันเกิด เพื่อความสบายใจสูงสุดของผู้ใช้งาน\n• หากไม่ระบุวันเกิด ระบบจะอนุโลมให้ใช้งานได้กับทุกวันเกิด",
                color = Color.LightGray,
                fontSize = 12.sp
            )

            Text(
                text = "2. หลักการจัดอันดับแบบ Hybrid",
                color = GoldPrimary,
                fontWeight = FontWeight.Bold,
                fontSize = 14.sp
            )
            Text(
                text = "• คะแนนเลขศาสตร์และความหมายถูกปรับให้อยู่บนสเกล 0–100 ก่อนถ่วงน้ำหนัก\n• คู่เลขศาสตร์และพลังเงามงคลสมบูรณ์แบบได้โบนัส +8 ที่แสดงชัดเจน\n• กรณีคะแนนเท่ากัน: ใช้คะแนนเลขศาสตร์รวมเป็นตัวตัดสิน\n• ไม่ใช้ทักษาหรือจำนวนพยางค์เป็นเงื่อนไข เพื่อคงตัวเลือกที่หลากหลาย",
                color = Color.LightGray,
                fontSize = 12.sp
            )

            Text(
                text = "3. หลักการให้ตัวเลือกที่หลากหลาย",
                color = GoldPrimary,
                fontWeight = FontWeight.Bold,
                fontSize = 14.sp
            )
            Text(
                text = "• ทุกชื่อแข่งขันกันจากคะแนนเลขศาสตร์และความหมาย โดยไม่ตัดชื่อเพราะจำนวนพยางค์\n• โหมดชื่อ+นามสกุลคำนวณผลรวมจากฐานข้อมูลกลาง เพื่อให้ผลลัพธ์ถูกต้องและตรวจสอบได้",
                color = Color.LightGray,
                fontSize = 12.sp
            )

            Spacer(modifier = Modifier.height(8.dp))

            Button(
                onClick = onDismiss,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary)
            ) {
                Text(
                    text = "เข้าใจแล้ว (ปิดคู่มือ)",
                    color = Color.Black,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}
