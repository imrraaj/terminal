import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { VisualizationTab } from "./components/tabs/VisualizationTab";
import { ActiveStrategiesTab } from "./components/tabs/ActiveStrategiesTab";
import { PortfolioTab } from "./components/tabs/PortfolioTab";

function App() {
    return (
        <main className="bg-background text-foreground">
            <Tabs defaultValue="visualization" className="flex-1 flex flex-col overflow-hidden">
                <div className="border-b bg-card px-6 py-3">
                    <div className="flex items-center justify-between">
                        <h1 className="text-xl font-bold tracking-tight">HyperTerminal</h1>
                        <TabsList className="flex space-x-2 bg-none">
                            <TabsTrigger value="visualization">Visualization</TabsTrigger>
                            <TabsTrigger value="active-strategies">Active Strategies</TabsTrigger>
                            <TabsTrigger value="portfolio">Portfolio</TabsTrigger>
                        </TabsList>
                    </div>
                </div>

                <div className="flex-1 overflow-auto">
                    <TabsContent value="visualization" className="h-full m-0">
                        <VisualizationTab />
                    </TabsContent>

                    <TabsContent value="active-strategies" className="h-full m-0">
                        <ActiveStrategiesTab />
                    </TabsContent>

                    <TabsContent value="portfolio" className="h-full m-0">
                        <PortfolioTab />
                    </TabsContent>
                </div>
            </Tabs>
        </main>
    );
}

export default App;
