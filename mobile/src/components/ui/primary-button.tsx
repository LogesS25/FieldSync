import { Pressable, Text } from 'react-native';

export type ButtonVariant = 'brand' | 'danger' | 'neutral';

interface PrimaryButtonProps {
  label: string;
  onPress: () => void;
  disabled?: boolean;
  variant?: ButtonVariant;
}

// Literal lookup, not string interpolation — see components/ui/badge.tsx for
// why (NativeWind's Tailwind JIT scanner needs literal classnames).
const VARIANT_STYLES: Record<ButtonVariant, string> = {
  brand: 'bg-brand-600',
  danger: 'bg-rose-600',
  neutral: 'bg-slate-200',
};

const VARIANT_TEXT: Record<ButtonVariant, string> = {
  brand: 'text-white',
  danger: 'text-white',
  neutral: 'text-slate-700',
};

export function PrimaryButton({ label, onPress, disabled, variant = 'brand' }: PrimaryButtonProps) {
  return (
    <Pressable
      className={`items-center rounded-lg py-3 disabled:opacity-50 ${VARIANT_STYLES[variant]}`}
      disabled={disabled}
      onPress={onPress}
    >
      <Text className={`font-semibold ${VARIANT_TEXT[variant]}`}>{label}</Text>
    </Pressable>
  );
}
