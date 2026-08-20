import { Text, TextInput, type TextInputProps, View } from 'react-native';

interface LabeledInputProps extends TextInputProps {
  label: string;
  error?: string;
}

export function LabeledInput({ label, error, ...inputProps }: LabeledInputProps) {
  return (
    <View className="mb-3">
      <Text className="mb-1 text-sm font-medium text-slate-700">{label}</Text>
      <TextInput
        className="rounded-lg border border-slate-300 bg-white px-4 py-3 text-slate-900"
        placeholderTextColor="#94a3b8"
        {...inputProps}
      />
      {error ? <Text className="mt-1 text-sm text-red-600">{error}</Text> : null}
    </View>
  );
}
