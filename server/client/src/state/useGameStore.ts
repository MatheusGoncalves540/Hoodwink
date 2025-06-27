import { create } from 'zustand'

interface GameState {
  nickname: string
  players: string[]
  setNickname: (name: string) => void
}

export const useGameStore = create<GameState>((set) => ({
  nickname: '',
  players: [],
  setNickname: (name) => set({ nickname: name }),
}))
