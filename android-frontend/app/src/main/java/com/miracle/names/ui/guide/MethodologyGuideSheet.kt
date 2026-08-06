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
                text = "2. หลักการจัดอันดับ 3 ชั้น (3-Tier Hybrid Ranking)",
                color = GoldPrimary,
                fontWeight = FontWeight.Bold,
                fontSize = 14.sp
            )
            Text(
                text = "• Tier 1 (ตรงเป้าหมาย 100%): คัดเลือกชื่อที่มีอักษรทักษาเน้นที่ผู้ใช้เลือก (เช่น ศรี)\n• Tier 2 (ทักษามงคลเสริม): หากชื่อตรงทักษาหลักหมด ระบบจะเติมชื่อที่มีทักษามงคลอื่นๆ (เดช, มนตรี, มูละ, อุตสาหะ, อายุ) ให้ครบ 100 รายชื่ออัตโนมัติ\n• Tier 3 (พลังเลขศาสตร์มงคล): คัดเลือกชื่อเกรดมงคล (D10, D8, D5) ปลอดกาลกิณี",
                color = Color.LightGray,
                fontSize = 12.sp
            )

            Text(
                text = "3. หลักการอนุโลม 2 พยางค์ขึ้นไป (Smart Flexibility Rule)",
                color = GoldPrimary,
                fontWeight = FontWeight.Bold,
                fontSize = 14.sp
            )
            Text(
                text = "• ชื่อไทยนิยมใช้ 2 พยางค์ขึ้นไป หากบังคับผลรวมชื่อ+นามสกุลเป็น D10 อย่างเดียว จะทำให้ตัวเลือกน้อยเกินไป ระบบจึงยืดหยุ่นให้ชื่อตัวเป็นเกรดมงคล (D) แล้วคำนวณผลรวมกับนามสกุลให้เห็น เพื่อรักษาความงามของภาษาไว้",
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
