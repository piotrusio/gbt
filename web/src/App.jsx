import { AppLayout } from '@/components/app-layout'
import { FlowExample } from '@/components/flow-example'

function App() {
  return (
    <AppLayout>
      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Welcome to OGBT</h2>
          <p className="text-muted-foreground">
            A modern app with React Flow, Tailwind CSS, and shadcn/ui
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          <div className="rounded-lg border bg-card p-4">
            <h3 className="font-semibold">React Flow</h3>
            <p className="text-sm text-muted-foreground">
              Interactive node-based diagrams
            </p>
          </div>
          <div className="rounded-lg border bg-card p-4">
            <h3 className="font-semibold">shadcn/ui</h3>
            <p className="text-sm text-muted-foreground">
              Beautiful, accessible components
            </p>
          </div>
          <div className="rounded-lg border bg-card p-4">
            <h3 className="font-semibold">Tailwind CSS</h3>
            <p className="text-sm text-muted-foreground">
              Utility-first styling
            </p>
          </div>
        </div>

        <div>
          <h3 className="text-xl font-semibold mb-4">React Flow Demo</h3>
          <FlowExample />
        </div>
      </div>
    </AppLayout>
  )
}

export default App
