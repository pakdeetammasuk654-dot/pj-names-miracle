package com.miracle.names.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import com.miracle.names.R

val DarkColorScheme = darkColorScheme(
    primary = GoldPrimary,
    onPrimary = Color.Black,
    secondary = PurpleAccent,
    background = PurpleDark,
    surface = CardBgDark,
    onSurface = Color.White
)

private val LightColorScheme = lightColorScheme(
    primary = PurpleHeader,
    onPrimary = Color.White,
    secondary = GoldPrimary,
    background = Color(0xFFF1EDF4),
    surface = Color(0xFFFFFCFF),
    surfaceVariant = Color(0xFFE8E1EB),
    onSurface = TextPrimary,
    onSurfaceVariant = Color(0xFF5F5662),
    outline = Color(0xFF827586),
    outlineVariant = Color(0xFFD4CBD7)
)

val SarabunFontFamily = FontFamily(
    Font(R.font.sarabun_regular, FontWeight.Normal),
    Font(R.font.sarabun_medium, FontWeight.Medium),
    Font(R.font.sarabun_semibold, FontWeight.SemiBold),
    Font(R.font.sarabun_bold, FontWeight.Bold)
)

private val SarabunTypography = Typography().withFontFamily(SarabunFontFamily)

private fun Typography.withFontFamily(fontFamily: FontFamily): Typography = copy(
    displayLarge = displayLarge.copy(fontFamily = fontFamily),
    displayMedium = displayMedium.copy(fontFamily = fontFamily),
    displaySmall = displaySmall.copy(fontFamily = fontFamily),
    headlineLarge = headlineLarge.copy(fontFamily = fontFamily),
    headlineMedium = headlineMedium.copy(fontFamily = fontFamily),
    headlineSmall = headlineSmall.copy(fontFamily = fontFamily),
    titleLarge = titleLarge.copy(fontFamily = fontFamily),
    titleMedium = titleMedium.copy(fontFamily = fontFamily),
    titleSmall = titleSmall.copy(fontFamily = fontFamily),
    bodyLarge = bodyLarge.copy(fontFamily = fontFamily),
    bodyMedium = bodyMedium.copy(fontFamily = fontFamily),
    bodySmall = bodySmall.copy(fontFamily = fontFamily),
    labelLarge = labelLarge.copy(fontFamily = fontFamily),
    labelMedium = labelMedium.copy(fontFamily = fontFamily),
    labelSmall = labelSmall.copy(fontFamily = fontFamily)
)

@Composable
fun MiracleNamesTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme

    MaterialTheme(
        colorScheme = colorScheme,
        typography = SarabunTypography,
        content = content
    )
}
