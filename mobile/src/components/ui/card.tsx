import type { PropsWithChildren } from 'react';
import { View } from 'react-native';

interface CardProps extends PropsWithChildren {
  className?: string;
}

export function Card({ children, className }: CardProps) {
  return (
    <View
      className={`rounded-2xl border border-slate-100 bg-white p-5 shadow-sm shadow-slate-200 ${className ?? ''}`}
    >
      {children}
    </View>
  );
}
