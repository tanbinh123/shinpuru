import { AppTheme, getSystemTheme } from '../theme/theme';
import { Guild, User } from '../lib/shinpuru-ts/src';

import LocalStorageUtil from '../util/localstorage';
import { ModalState } from '../hooks/useModal';
import { Notification } from '../components/Notifications';
import create from 'zustand';

export type FetchLocked<T> = {
  value: T | undefined;
  isFetching: boolean;
};

export interface Store {
  theme: AppTheme;
  setTheme: (v: AppTheme) => void;

  accentColor?: string;
  setAccentColor: (v?: string) => void;

  selfUser: FetchLocked<User>;
  setSelfUser: (selfUser: Partial<FetchLocked<User>>) => void;

  guilds?: Guild[];
  setGuilds: (guilds?: Guild[]) => void;

  selectedGuild?: Guild;
  setSelectedGuild: (selectedGuild: Guild) => void;

  notifications: Notification[];
  setNotifications: (notifications: Notification[]) => void;

  modal: ModalState<any>;
  setModal: (modal: ModalState<any>) => void;
}

export const useStore = create<Store>((set, get) => ({
  theme: LocalStorageUtil.get('shnp.theme', getSystemTheme())!,
  setTheme: (theme) => {
    set({ theme });
    LocalStorageUtil.set('shnp.theme', theme);
  },

  accentColor: LocalStorageUtil.get('shnp.accentcolor'),
  setAccentColor: (accentColor) => {
    set({ accentColor });
    if (accentColor === undefined) LocalStorageUtil.del('shnp.accentcolor');
    else LocalStorageUtil.set('shnp.accentcolor', accentColor);
  },

  selfUser: { value: undefined, isFetching: false },
  setSelfUser: (selfUser: Partial<FetchLocked<User>>) =>
    set({ selfUser: { ...get().selfUser, ...selfUser } }),

  guilds: undefined,
  setGuilds: (guilds) => set({ guilds }),

  selectedGuild: undefined,
  setSelectedGuild: (selectedGuild) => {
    set({ selectedGuild });
    if (!!selectedGuild) LocalStorageUtil.set('shnp.selectedguild', selectedGuild.id);
  },

  notifications: [],
  setNotifications: (notifications) => set({ notifications }),

  modal: { isOpen: false },
  setModal: (modal) => set({ modal }),
}));
