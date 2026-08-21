import { Ionicons } from '@expo/vector-icons';
import type { PropsWithChildren, ReactNode } from 'react';
import { Text, View } from 'react-native';

interface PageHeaderProps extends PropsWithChildren {
  title: string;
  description?: string;
  icon?: keyof typeof Ionicons.glyphMap;
  action?: ReactNode;
}

// Consistent page-top treatment: icon chip + title + one-line description,
// optionally with a trailing action (e.g. a filter pill). The Drawer's own
// top bar already carries the route name, so this is the richer in-body
// header every real (non-placeholder) screen uses instead of ad hoc
// text-2xl blocks.
export function PageHeader({ title, description, icon, action, children }: PageHeaderProps) {
  return (
    <View className="mb-6 flex-row items-start justify-between gap-3">
      <View className="flex-1 flex-row items-start gap-3">
        {icon ? (
          <View className="mt-0.5 h-10 w-10 items-center justify-center rounded-xl bg-brand-50">
            <Ionicons name={icon} size={19} color="#4f46e5" />
          </View>
        ) : null}
        <View className="flex-1">
          <Text className="text-xl font-bold text-slate-900">{title}</Text>
          {description ? <Text className="mt-1 text-sm leading-5 text-slate-500">{description}</Text> : null}
          {children}
        </View>
      </View>
      {action}
    </View>
  );
}
