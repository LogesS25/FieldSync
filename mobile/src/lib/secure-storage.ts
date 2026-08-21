import { Platform } from 'react-native';
import * as SecureStore from 'expo-secure-store';
import type { StateStorage } from 'zustand/middleware';

// Adapts expo-secure-store (device Keychain/Keystore) to zustand's
// StateStorage interface so persisted auth state never sits in plain
// AsyncStorage on native. expo-secure-store has no web implementation at
// all (its web module is a literal empty object), so calling it on web
// throws — which previously left zustand's persist rehydration promise
// unresolved forever, hanging the whole app behind an infinite loading
// spinner before a single API call was ever made. Web has no secure-enclave
// equivalent, so it falls back to localStorage (same trust level as any
// other web app's session storage, and this is a dev-only build target).
export const secureStorage: StateStorage =
  Platform.OS === 'web'
    ? {
        getItem: (name) => (typeof localStorage === 'undefined' ? null : localStorage.getItem(name)),
        setItem: (name, value) => {
          if (typeof localStorage !== 'undefined') localStorage.setItem(name, value);
        },
        removeItem: (name) => {
          if (typeof localStorage !== 'undefined') localStorage.removeItem(name);
        },
      }
    : {
        getItem: async (name) => (await SecureStore.getItemAsync(name)) ?? null,
        setItem: (name, value) => SecureStore.setItemAsync(name, value),
        removeItem: (name) => SecureStore.deleteItemAsync(name),
      };
