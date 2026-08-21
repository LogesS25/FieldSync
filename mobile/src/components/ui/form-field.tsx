import type { PropsWithChildren } from 'react';
import { Text, View } from 'react-native';

interface FormFieldProps extends PropsWithChildren {
  label: string;
  error?: string;
  hint?: string;
}

// Label + control + error/hint wrapper for non-TextInput fields (pill
// pickers, date pickers, file pickers) — LabeledInput covers the TextInput
// case, this covers everything else, so every field in the app gets the
// same label/spacing/error treatment regardless of control type.
export function FormField({ label, error, hint, children }: FormFieldProps) {
  return (
    <View className="mb-4">
      <Text className="mb-2 text-sm font-medium text-slate-700">{label}</Text>
      {children}
      {error ? (
        <Text className="mt-1.5 text-sm text-red-600">{error}</Text>
      ) : hint ? (
        <Text className="mt-1.5 text-sm text-slate-400">{hint}</Text>
      ) : null}
    </View>
  );
}
