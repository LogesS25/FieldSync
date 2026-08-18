import { Text, View } from 'react-native';

export type BadgeTone = 'success' | 'warning' | 'danger' | 'neutral' | 'info';

interface BadgeProps {
  label: string;
  tone?: BadgeTone;
}

// Tone → className is a literal lookup (not string interpolation) because
// NativeWind's Tailwind JIT scanner only picks up classnames it can see as
// literal strings in source — a template-built `bg-${tone}-100` would
// silently not exist in the compiled stylesheet.
const TONE_STYLES: Record<BadgeTone, { container: string; text: string }> = {
  success: { container: 'bg-emerald-50 border-emerald-200', text: 'text-emerald-700' },
  warning: { container: 'bg-amber-50 border-amber-200', text: 'text-amber-700' },
  danger: { container: 'bg-rose-50 border-rose-200', text: 'text-rose-700' },
  neutral: { container: 'bg-slate-100 border-slate-200', text: 'text-slate-600' },
  info: { container: 'bg-brand-50 border-brand-200', text: 'text-brand-700' },
};

export function Badge({ label, tone = 'neutral' }: BadgeProps) {
  const styles = TONE_STYLES[tone];
  return (
    <View className={`self-start rounded-full border px-2.5 py-1 ${styles.container}`}>
      <Text className={`text-xs font-medium ${styles.text}`}>{label}</Text>
    </View>
  );
}
