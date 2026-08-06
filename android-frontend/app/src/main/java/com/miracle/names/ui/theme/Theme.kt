package com.miracle.names.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val DarkColorScheme = darkColorScheme(
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

@Composable
fun MiracleNamesTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}
