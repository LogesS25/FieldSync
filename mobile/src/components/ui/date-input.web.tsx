import { createElement } from 'react';
import { Text, View } from 'react-native';

interface DateInputProps {
  label: string;
  value: string; // YYYY-MM-DD
  onChange: (value: string) => void;
  error?: string;
}

// react-native-web renders inside the DOM, so on web we use the browser's
// native <input type="date"> (calendar picker built in) instead of
// reimplementing one. `createElement` with a string tag sidesteps
// react-native's JSX.IntrinsicElements typings, which don't include raw DOM
// elements — this file only ever runs on web (Metro resolves `.web.tsx`
// over `.tsx` for web builds), so that's the whole point of it existing.
export function DateInput({ label, value, onChange, error }: DateInputProps) {
  return (
    <View className="mb-3">
      <Text className="mb-1 text-sm font-medium text-slate-700">{label}</Text>
      {createElement('input', {
        type: 'date',
        value,
        onChange: (e: { target: { value: string } }) => onChange(e.target.value),
        style: {
          border: '1px solid #cbd5e1',
          borderRadius: 8,
          padding: '11px 14px',
          fontSize: 16,
          color: '#0f172a',
          backgroundColor: 'white',
          width: '100%',
          fontFamily: 'inherit',
        },
      })}
      {error ? <Text className="mt-1 text-sm text-red-600">{error}</Text> : null}
    </View>
  );
}
