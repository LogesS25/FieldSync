import * as SecureStore from 'expo-secure-store';
import type { StateStorage } from 'zustand/middleware';

// Adapts expo-secure-store (device Keychain/Keystore) to zustand's
// StateStorage interface so persisted auth state never sits in plain
// AsyncStorage.
export const secureStorage: StateStorage = {
  getItem: async (name) => (await SecureStore.getItemAsync(name)) ?? null,
  setItem: (name, value) => SecureStore.setItemAsync(name, value),
  removeItem: (name) => SecureStore.deleteItemAsync(name),
};
