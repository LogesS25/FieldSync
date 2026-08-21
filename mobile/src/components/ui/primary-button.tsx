import { ActivityIndicator, Pressable, Text } from 'react-native';

export type ButtonVariant = 'brand' | 'danger' | 'neutral' | 'outline';

interface PrimaryButtonProps {
  label: string;
  onPress: () => void;
  disabled?: boolean;
  loading?: boolean;
  variant?: ButtonVariant;
}

// Literal lookup, not string interpolation — see components/ui/badge.tsx for
// why (NativeWind's Tailwind JIT scanner needs literal classnames).
const VARIANT_STYLES: Record<ButtonVariant, string> = {
  brand: 'bg-brand-600 active:bg-brand-700',
  danger: 'bg-rose-600 active:bg-rose-700',
  neutral: 'bg-slate-100 active:bg-slate-200',
  outline: 'border border-slate-300 bg-white active:bg-slate-50',
};

const VARIANT_TEXT: Record<ButtonVariant, string> = {
  brand: 'text-white',
  danger: 'text-white',
  neutral: 'text-slate-700',
  outline: 'text-slate-700',
};

const SPINNER_COLOR: Record<ButtonVariant, string> = {
  brand: '#ffffff',
  danger: '#ffffff',
  neutral: '#334155',
  outline: '#334155',
};

export function PrimaryButton({ label, onPress, disabled, loading, variant = 'brand' }: PrimaryButtonProps) {
  return (
    <Pressable
      className={`flex-row items-center justify-center gap-2 rounded-xl py-3.5 disabled:opacity-50 ${VARIANT_STYLES[variant]}`}
      disabled={disabled || loading}
      onPress={onPress}
    >
      {loading ? <ActivityIndicator size="small" color={SPINNER_COLOR[variant]} /> : null}
      <Text className={`font-semibold ${VARIANT_TEXT[variant]}`}>{label}</Text>
    </Pressable>
  );
}
