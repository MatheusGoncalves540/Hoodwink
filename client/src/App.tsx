import { useSocket } from './hooks/useSocket'
import { useGameStore } from './state/useGameStore'

function App() {
  useSocket('ws://localhost:8080/ws')

  const nickname = useGameStore((state) => state.nickname)
  const setNickname = useGameStore((state) => state.setNickname)

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100">
      <h1 className="text-2xl font-bold mb-4">Hoodwink</h1>
      <input
        className="p-2 border rounded mb-2"
        placeholder="Digite seu nickname"
        value={nickname}
        onChange={(e) => setNickname(e.target.value)}
      />
      <p className="text-sm text-gray-600">Nickname: {nickname}</p>
    </div>
  )
}

export default App
