export async function initOrchestration(terminal) {
    const container = document.getElementById('d3-canvas');
    const width = container.clientWidth;
    const height = container.clientHeight;

    d3.select(container).selectAll('*').remove();

    const svg = d3.select(container)
        .append('svg')
        .attr('width', '100%')
        .attr('height', '100%')
        .attr('viewBox', '0 0 6000 10000') // Scale large graph to fit
        .attr('preserveAspectRatio', 'xMidYMid meet');

    const g = svg.append('g');

    const zoom = d3.zoom()
        .scaleExtent([0.05, 8]) // Allow much deeper zoom and wider view
        .on('zoom', (event) => {
            g.attr('transform', event.transform);
            updateLabels(event.transform.k);
        });

    svg.call(zoom);

    function updateLabels(scale) {
        // Show labels only when zoomed in enough or on hover
        g.selectAll('.node text')
            .attr('opacity', d => {
                // Always show current node label
                if (currentScenario && d.name === currentScenario.path[currentStepIndex]) return 1;
                // Show all labels if zoomed in (lowered threshold for better default view)
                if (scale > 0.3) return 0.7;
                // Otherwise show dimmed
                return 0.4;
            });
    }

    // State
    let architectureData = null;
    let currentScenario = null;
    let isPlaying = false;
    let currentStepIndex = 0;
    let accumulatedCost = 0;
    let accumulatedTime = 0;
    let timerInterval = null;
    let sankeyGraph = null;
    let nodeMap = new Map();
    let scenariosData = null;

    // Global data for layout
    let graphNodes = [];
    let graphLinks = [];
    let nodeDepthMap = new Map();
    let graphMaxDepth = 0;

    // UI Elements
    const statModel = document.getElementById('stat-model');
    const statCost = document.getElementById('stat-cost');
    const statTime = document.getElementById('stat-time');
    const statStatus = document.getElementById('stat-status');

    // Load Data
    try {
        const timestamp = new Date().getTime();
        const archResponse = await fetch(`architecture.json?t=${timestamp}`);
        architectureData = await archResponse.json();

        const scenariosResponse = await fetch(`scenarios.json?t=${timestamp}`);
        scenariosData = await scenariosResponse.json();

        buildArchitectureGraph();
    } catch (e) {
        console.error("Failed to load data", e);
    }

    // Event Listeners
    document.addEventListener('terminal-command-complete', () => {
        if (scenariosData && scenariosData.auth) {
            loadScenario(scenariosData.auth);
        }
    });

    function loadScenario(scenario) {
        console.log('Loading scenario:', scenario.title);
        currentScenario = scenario;
        resetScenario();
        play();
    }

    function buildArchitectureGraph() {
        if (!architectureData) {
            return;
        }

        // Build full graph from architecture.json
        const nodesSet = new Set();
        const rawLinks = [];

        architectureData.forEach(([source, target, value]) => {
            nodesSet.add(source);
            nodesSet.add(target);
            rawLinks.push({ source, target, value });
        });

        const nodes = Array.from(nodesSet).map(name => ({ name }));

        // Cycle removal for Sankey
        const adj = new Map();
        rawLinks.forEach(l => {
            if (!adj.has(l.source)) adj.set(l.source, []);
            adj.get(l.source).push(l.target);
        });

        const visited = new Set();
        const recursionStack = new Set();
        const acyclicLinks = [];
        const keptLinks = new Set();

        function isCyclic(node) {
            visited.add(node);
            recursionStack.add(node);

            const neighbors = adj.get(node) || [];
            for (const neighbor of neighbors) {
                const edgeExists = rawLinks.some(l => l.source === node && l.target === neighbor);
                if (!edgeExists) continue;

                if (!visited.has(neighbor)) {
                    isCyclic(neighbor);
                    const link = rawLinks.find(l => l.source === node && l.target === neighbor);
                    if (link && !keptLinks.has(link)) {
                        acyclicLinks.push(link);
                        keptLinks.add(link);
                    }
                } else if (recursionStack.has(neighbor)) {
                    console.warn(`Cycle: ${node} -> ${neighbor}`);
                } else {
                    const link = rawLinks.find(l => l.source === node && l.target === neighbor);
                    if (link && !keptLinks.has(link)) {
                        acyclicLinks.push(link);
                        keptLinks.add(link);
                    }
                }
            }

            recursionStack.delete(node);
            return false;
        }

        nodes.forEach(node => {
            if (!visited.has(node.name)) {
                isCyclic(node.name);
            }
        });

        console.log(`Architecture: ${nodes.length} nodes, ${acyclicLinks.length} links`);

        // CALCULATE LOGICAL LAYERS FOR ALL NODES
        // BFS from input nodes to determine depth
        const nodeDepth = new Map();
        const inputNodes = ['UserInput', 'CLI_Main', 'CommandParser'];
        const queue = inputNodes.map(name => ({ name, depth: 0 }));
        const visitedBFS = new Set();

        // Use raw links for depth calculation to ensure connectivity
        const allLinks = [];
        architectureData.forEach(([source, target, value]) => {
            allLinks.push({ source, target });
        });

        let maxDepth = 0;

        while (queue.length > 0) {
            const { name, depth } = queue.shift();
            if (visitedBFS.has(name)) continue;
            visitedBFS.add(name);

            nodeDepth.set(name, depth);
            maxDepth = Math.max(maxDepth, depth);

            // Find all outgoing links
            const outgoing = allLinks.filter(l => l.source === name);
            outgoing.forEach(link => {
                if (!visitedBFS.has(link.target)) {
                    // Special handling for Tools - keep them closer
                    const isTool = link.target.includes('Tools_') ||
                        link.target.includes('Tool') ||
                        link.target.includes('_file') ||
                        link.target.includes('_command') ||
                        link.target.includes('_patch') ||
                        link.target.includes('_map');
                    const depthIncrement = isTool ? 0.1 : 1;
                    queue.push({ name: link.target, depth: depth + depthIncrement });
                }
            });
        }

        // Fallback for unvisited nodes
        let changed = true;
        while (changed) {
            changed = false;
            nodes.forEach(node => {
                if (!nodeDepth.has(node.name)) {
                    const parentLink = allLinks.find(l => l.target === node.name && nodeDepth.has(l.source));
                    if (parentLink) {
                        const parentDepth = nodeDepth.get(parentLink.source);
                        const isTool = node.name.includes('Tools_') ||
                            node.name.includes('Tool') ||
                            node.name.includes('_file') ||
                            node.name.includes('_command') ||
                            node.name.includes('_patch') ||
                            node.name.includes('_map');
                        const depthIncrement = isTool ? 0.1 : 1;
                        nodeDepth.set(node.name, parentDepth + depthIncrement);
                        changed = true;
                    }
                }
            });
        }

        // 3. Filter links to ensure strict DAG (no back-edges or same-depth edges)
        const strictAcyclicLinks = [];
        let removedLinks = 0;

        acyclicLinks.forEach(link => {
            const sourceDepth = nodeDepth.get(link.source);
            const targetDepth = nodeDepth.get(link.target);

            // Only allow links that go forward in depth
            if (targetDepth > sourceDepth) {
                strictAcyclicLinks.push(link);
            } else {
                removedLinks++;
                // console.log(`[DAG] Removing cycle/back-edge: ${link.source} (${sourceDepth}) -> ${link.target} (${targetDepth})`);
            }
        });

        console.log(`[DAG] Removed ${removedLinks} back-edges to ensure strict acyclic graph.`);

        // Store prepared data globally
        graphNodes = nodes;
        graphLinks = strictAcyclicLinks;
        nodeDepthMap = nodeDepth;
        graphMaxDepth = maxDepth;

        // Initial layout
        updateLayout();
    }

    function updateLayout() {
        try {
            const sankeyWidth = 6000;
            const sankeyHeight = 10000;

            // Use original architecture graph - NO unrolling
            const renderNodes = graphNodes.map(d => Object.assign({}, d));
            const renderLinks = graphLinks.map(d => Object.assign({}, d));

            console.log('[DEBUG] renderNodes count:', renderNodes.length);
            console.log('[DEBUG] renderLinks count:', renderLinks.length);
            // Custom Align Function - use BFS depth for ALL nodes
            const sankey = d3.sankey()
                .nodeWidth(150)  // Increased from 80 to accommodate full names
                .nodePadding(60)
                .extent([[100, 100], [sankeyWidth - 100, sankeyHeight - 100]])
                .nodeId(d => d.name)
                .nodeAlign(d3.sankeyCenter); // Center nodes for balanced "diamond-like" layout

            sankeyGraph = sankey({
                nodes: renderNodes,
                links: renderLinks
            });

            sankeyGraph.nodes.forEach(node => {
                nodeMap.set(node.name, node);
                node.color = getNodeColor(node.name);

                // Safety: Ensure coordinates are numbers
                if (isNaN(node.x0)) node.x0 = 0;
                if (isNaN(node.x1)) node.x1 = 10;
                if (isNaN(node.y0)) node.y0 = 0;
                if (isNaN(node.y1)) node.y1 = 10;
            });

            console.log("Layout updated with", sankeyGraph.nodes.length, "nodes");

            try {
                renderSankey();
            } catch (renderError) {
                console.error("renderSankey failed:", renderError);
            }
            // Initial zoom - show full graph
            const bounds = g.node().getBBox();

            if (bounds.width > 0 && bounds.height > 0) {
                const scale = Math.min(width / bounds.width, height / bounds.height) * 0.75;
                const tx = (width - bounds.width * scale) / 2 - bounds.x * scale;
                const ty = (height - bounds.height * scale) / 2 - bounds.y * scale;

                if (!isNaN(scale) && !isNaN(tx) && !isNaN(ty)) {
                    svg.call(zoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
                }
            }

        } catch (e) {
            console.error("Sankey failed during layout:", e);
            const errorDiv = document.createElement('div');
            errorDiv.style.position = 'absolute';
            errorDiv.style.top = '10px';
            errorDiv.style.left = '10px';
            errorDiv.style.backgroundColor = 'red';
            errorDiv.style.color = 'white';
            errorDiv.style.padding = '20px';
            errorDiv.style.zIndex = '9999';
            errorDiv.innerText = "Error: " + e.message + "\n" + e.stack;
            document.body.appendChild(errorDiv);
        }
    }

    function buildScenarioGraph() {
        if (!currentScenario) return;

        console.log('Building graph for scenario path...');

        // Create nodes from scenario path only
        const pathNodes = currentScenario.path.map((name, index) => ({
            name,
            index,
            color: getNodeColor(name)
        }));

        // Create sequential links
        const pathLinks = [];
        for (let i = 0; i < pathNodes.length - 1; i++) {
            pathLinks.push({
                source: i,
                target: i + 1,
                value: 5
            });
        }

        console.log(`Scenario graph: ${pathNodes.length} nodes, ${pathLinks.length} links`);

        // Sankey Layout - horizontal sequential flow
        const sankeyWidth = pathNodes.length * 180;
        const sankeyHeight = 600;

        const sankey = d3.sankey()
            .nodeWidth(25)
            .nodePadding(80)
            .extent([[50, 50], [sankeyWidth, sankeyHeight]])
            .nodeId(d => d.index);

        try {
            sankeyGraph = sankey({
                nodes: pathNodes.map(d => Object.assign({}, d)),
                links: pathLinks.map(d => Object.assign({}, d))
            });

            // Map nodes by name for lookup
            sankeyGraph.nodes.forEach(node => {
                nodeMap.set(node.name, node);
                if (isNaN(node.x0)) node.x0 = 0;
                if (isNaN(node.x1)) node.x1 = 10;
                if (isNaN(node.y0)) node.y0 = 0;
                if (isNaN(node.y1)) node.y1 = 10;
            });

            renderSankey();

            // Initial zoom to fit
            const bounds = g.node().getBBox();
            if (bounds.width > 0 && bounds.height > 0) {
                const scale = Math.min(width / bounds.width, height / bounds.height) * 0.9;
                const tx = (width - bounds.width * scale) / 2 - bounds.x * scale;
                const ty = (height - bounds.height * scale) / 2 - bounds.y * scale;

                if (!isNaN(scale) && !isNaN(tx) && !isNaN(ty)) {
                    svg.call(zoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
                }
            }

        } catch (e) {
            console.error("Sankey failed", e);
        }
    }

    function getNodeColor(name) {
        if (name.includes('CLI') || name.includes('Command')) return '#94a3b8';
        if (name.includes('Intelligence') || name.includes('Model') || name.includes('ML')) return '#8b5cf6';
        if (name.includes('Agent')) {
            if (name.includes('Analyzer')) return '#3b82f6';
            if (name.includes('Planner')) return '#6366f1';
            if (name.includes('Editor')) return '#10b981';
            if (name.includes('Validator')) return '#f59e0b';
            return '#0ea5e9';
        }
        if (name.includes('Tools') || name.includes('File')) return '#64748b';
        if (name.includes('Success') || name.includes('Output')) return '#22c55e';
        if (name.includes('Failure') || name.includes('Retry')) return '#ef4444';
        return '#cbd5e1';
    }

    function renderSankey() {
        console.log('[DEBUG] renderSankey called');
        if (!sankeyGraph || !sankeyGraph.nodes || !sankeyGraph.links) {
            console.error('[DEBUG] Invalid sankeyGraph in renderSankey');
            return;
        }
        console.log('[DEBUG] Rendering', sankeyGraph.nodes.length, 'nodes and', sankeyGraph.links.length, 'links');

        g.selectAll('*').remove();

        if (sankeyGraph.nodes.length === 0) {
            g.append('text')
                .attr('x', 3000)
                .attr('y', 5000)
                .attr('text-anchor', 'middle')
                .attr('font-size', '100px')
                .attr('fill', 'red')
                .text('Graph Empty');
            return;
        }

        // Links
        const links = g.append('g')
            .attr('fill', 'none')
            .selectAll('path')
            .data(sankeyGraph.links)
            .join('path')
            .attr('class', 'link')
            .attr('d', d3.sankeyLinkHorizontal())
            .attr('stroke', '#cbd5e1')
            .attr('stroke-width', d => Math.max(2, d.width))
            .attr('stroke-opacity', 0.05);

        // Nodes
        const nodes = g.append('g')
            .selectAll('g')
            .data(sankeyGraph.nodes)
            .join('g')
            .attr('class', 'node')
            .attr('transform', d => `translate(${d.x0},${d.y0})`);

        nodes.append('rect')
            .attr('height', d => d.y1 - d.y0)
            .attr('width', d => d.x1 - d.x0)
            .attr('fill', d => d.color)
            .attr('fill-opacity', 0.6)  // Semi-transparent for lighter look
            .attr('rx', 3)
            .attr('stroke', '#94a3b8')  // Softer border color
            .attr('stroke-width', 2)
            .attr('opacity', 0.15);  // Overall dimmed by default

        // Labels positioned ABOVE nodes (Transformer Explainer style)
        nodes.append('text')
            .attr('x', d => (d.x1 - d.x0) / 2)
            .attr('y', -8)  // Position above the rectangle
            .attr('text-anchor', 'middle')
            .attr('font-family', 'Inter, sans-serif')
            .attr('font-size', '11px')
            .attr('font-weight', '500')
            .attr('fill', '#1e293b')
            .attr('opacity', 0.7) // Dimmed by default but visible
            .text(d => d.name);

        // Add hover behavior
        nodes.on('mouseenter', function (event, d) {
            // Brighten node and label on hover
            d3.select(this).select('rect').transition().duration(200).attr('opacity', 0.3);
            d3.select(this).select('text').transition().duration(200).attr('opacity', 1);
        }).on('mouseleave', function (event, d) {
            const currentScale = d3.zoomTransform(svg.node()).k;
            const isCurrent = currentScenario && d.name === currentScenario.path[currentStepIndex];
            if (!isCurrent && currentScale <= 0.8) {
                d3.select(this).select('rect').transition().duration(200).attr('opacity', 0.15);
                d3.select(this).select('text').transition().duration(200).attr('opacity', 0.7);
            }
        });

        console.log("Rendered:", nodes.size(), "nodes");
    }

    function play() {
        if (!currentScenario) return;
        isPlaying = true;
        statStatus.textContent = 'RUNNING';
        statStatus.className = 'stat-value status-badge running';
        runStep(currentStepIndex);
    }

    function pause() {
        isPlaying = false;
        statStatus.textContent = 'PAUSED';
        statStatus.className = 'stat-value status-badge';
        clearTimeout(timerInterval);
    }

    function resetScenario() {
        pause();
        currentStepIndex = 0;
        accumulatedCost = 0;
        accumulatedTime = 0;

        // Reset visuals
        g.selectAll('.node rect')
            .attr('opacity', 0.05)
            .attr('stroke', '#fff')
            .attr('stroke-width', 2);

        g.selectAll('.link')
            .attr('stroke-opacity', 0.05)
            .attr('stroke', '#cbd5e1');

        updateStats();
    }

    function updateStats(model, cost, time, status) {
        // Handle undefined values
        const safeCost = (cost !== undefined && cost !== null) ? cost : accumulatedCost;
        const safeTime = (time !== undefined && time !== null) ? time : accumulatedTime;

        if (model) statModel.textContent = model;
        statCost.textContent = '$' + safeCost.toFixed(4);
        statTime.textContent = (safeTime / 1000).toFixed(1) + 's';  // Convert ms to s
        if (status) {
            statStatus.textContent = status;
            statStatus.className = 'stat-value status-badge' + (status === 'RUNNING' ? ' running' : status === 'COMPLETED' ? ' success' : '');
        }
    }

    function runStep(index) {
        if (!currentScenario || index >= currentScenario.steps.length) {
            finishScenario();
            return;
        }

        // Skip to CommandParser if we're at the beginning
        if (index < 2 && currentScenario.path[2] === 'CommandParser') {
            currentStepIndex = 2;
            index = 2;
        }
        const step = currentScenario.steps[index];
        // Use original node name (no unique ID)
        const currentNodeName = step.node;

        // Log to terminal
        if (terminal) {
            const costStr = step.cost > 0 ? ` ($${step.cost.toFixed(5)})` : '';
            terminal.log(`${currentNodeName} [${step.model}]${costStr}`, 'system');
        }

        // Update Stats
        accumulatedCost += step.cost;
        accumulatedTime += step.duration;
        updateStats();

        // Highlight Node
        g.selectAll('.node rect')
            .transition()
            .duration(300)
            .attr('opacity', d => d.name === currentNodeName ? 0.8 : 0.15)  // Active node more opaque
            .attr('stroke', d => d.name === currentNodeName ? '#1e293b' : '#94a3b8')  // Darker stroke for active
            .attr('stroke-width', d => d.name === currentNodeName ? 3 : 2);

        // Highlight Labels
        g.selectAll('.node text')
            .transition()
            .duration(300)
            .attr('opacity', d => d.name === currentNodeName ? 1 : 0.5)  // Full opacity for current
            .attr('font-weight', d => d.name === currentNodeName ? '700' : '500');  // Bold for current

        // Highlight Links (Path)
        g.selectAll('.link')
            .transition()
            .duration(300)
            .attr('stroke-opacity', d => {
                if (d.source.name === currentNodeName || d.target.name === currentNodeName) return 0.6;
                return 0.05;
            })
            .attr('stroke', d => {
                if (d.source.name === currentNodeName) return '#22c55e'; // Outgoing
                if (d.target.name === currentNodeName) return '#3b82f6'; // Incoming
                return '#cbd5e1';
            });

        // Auto-scroll/Zoom to current node
        const currentNode = nodeMap.get(currentNodeName);
        if (currentNode) {
            // Force zoom to 3.5x for clarity
            const targetScale = 3.5;

            // Center on node
            const tx = (width / 2) - ((currentNode.x0 + currentNode.x1) / 2) * targetScale;
            const ty = (height / 2) - ((currentNode.y0 + currentNode.y1) / 2) * targetScale;

            svg.transition()
                .duration(500)
                .call(zoom.transform, d3.zoomIdentity.translate(tx, ty).scale(targetScale));

            // Show label for current node
            g.selectAll('.node text')
                .filter(d => d.name === currentNodeName)
                .transition()
                .attr('opacity', 1);
        }

        // Fog of War - show path nodes and neighbors
        const visibleNodes = new Set();
        // Add all path nodes
        currentScenario.path.forEach(n => visibleNodes.add(n));

        // Add immediate neighbors of current node
        if (currentNode) {
            sankeyGraph.links.forEach(l => {
                if (l.source.name === currentNodeName) visibleNodes.add(l.target.name);
                if (l.target.name === currentNodeName) visibleNodes.add(l.source.name);
            });
        }

        g.selectAll('.node rect')
            .filter(d => visibleNodes.has(d.name))
            .transition()
            .duration(300)
            .attr('opacity', d => d.name === currentNodeName ? 0.9 : 0.2);

        g.selectAll('.node text')
            .filter(d => visibleNodes.has(d.name))
            .transition()
            .duration(500)
            .attr('opacity', d => d.name === currentNodeName ? 1 : 0.7);

        updateStats(step.model, accumulatedCost, accumulatedTime, 'RUNNING');

        // Next Step
        currentStepIndex++;
        if (isPlaying) {
            timerInterval = setTimeout(() => {
                runStep(currentStepIndex);
            }, step.duration * 4); // Slowed down by 25% (3 to 4)
        }
    }

    function finishScenario() {
        isPlaying = false;
        statStatus.textContent = 'COMPLETED';
        statStatus.className = 'stat-value status-badge success';

        // Reveal all nodes and links
        console.log("Revealing all nodes...");
        g.selectAll('.node rect')
            .transition()
            .duration(1000)
            .attr('opacity', 0.6)
            .attr('stroke-width', 2);

        g.selectAll('.node text')
            .transition()
            .duration(1000)
            .attr('opacity', 0.8);

        g.selectAll('.link')
            .transition()
            .duration(1000)
            .attr('stroke-opacity', 0.3);

        // Zoom out to show full graph
        setTimeout(() => {
            const bounds = g.node().getBBox();
            if (bounds.width > 0 && bounds.height > 0) {
                const scale = Math.min(width / bounds.width, height / bounds.height) * 0.85;
                const tx = (width - bounds.width * scale) / 2 - bounds.x * scale;
                const ty = (height - bounds.height * scale) / 2 - bounds.y * scale;

                svg.transition()
                    .duration(1500)
                    .call(zoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
            }
        }, 500);
    }
}
