'use client'

interface Tab {
  key: string
  label: string
}

export function Tabs({ tabs, activeKey, onChange }: { tabs: Tab[]; activeKey: string; onChange: (key: string) => void }) {
  return (
    <div className="flex border-b border-gray-200">
      {tabs.map((tab) => (
        <button
          key={tab.key}
          onClick={() => onChange(tab.key)}
          className={`border-b-2 px-4 py-2 text-sm font-medium ${
            tab.key === activeKey
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700'
          }`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
