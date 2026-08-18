import type { PropsWithChildren } from 'react';
import { ScrollView, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

interface ScreenContainerProps extends PropsWithChildren {
  scroll?: boolean;
}

// Shared page shell: safe-area aware, consistent horizontal rhythm, one
// background color for the whole app. Every real screen should sit inside
// this instead of reinventing padding/background per screen.
export function ScreenContainer({ scroll = false, children }: ScreenContainerProps) {
  const Wrapper = scroll ? ScrollView : View;

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={['top', 'left', 'right']}>
      <Wrapper className={scroll ? undefined : 'flex-1'} contentContainerClassName={scroll ? 'px-5 py-6' : undefined}>
        {scroll ? children : <View className="flex-1 px-5 py-6">{children}</View>}
      </Wrapper>
    </SafeAreaView>
  );
}
