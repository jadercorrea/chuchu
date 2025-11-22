let allModels = [];
let filteredModels = [];
let selectedModels = [];
const MAX_COMPARISON = 4;

async function init() {
    try {
        const response = await fetch('/assets/data/models.json');
        const data = await response.json();
        allModels = data.models;
        
        document.getElementById('last-updated').textContent = 
            new Date(data.metadata.generated_at).toLocaleDateString();
        
        applyFiltersAndSort();
        setupEventListeners();
        calculateCosts();
    } catch (error) {
        console.error('Failed to load models:', error);
        document.getElementById('model-grid').innerHTML = 
            '<p>Failed to load models. Please try again later.</p>';
    }
}

function setupEventListeners() {
    document.querySelectorAll('.filters input[type="checkbox"]').forEach(checkbox => {
        checkbox.addEventListener('change', applyFiltersAndSort);
    });
    
    document.getElementById('sort-select').addEventListener('change', applyFiltersAndSort);
    
    document.getElementById('clear-comparison').addEventListener('click', () => {
        selectedModels = [];
        document.getElementById('comparison-section').style.display = 'none';
        renderModels();
    });
    
    document.getElementById('requests-per-day').addEventListener('input', calculateCosts);
    document.getElementById('tokens-per-request').addEventListener('input', calculateCosts);
}

function applyFiltersAndSort() {
    const providerFilters = Array.from(document.querySelectorAll('.filters .checkbox-group input[value]:checked'))
        .filter(cb => ['groq', 'openrouter', 'ollama', 'openai', 'deepseek'].includes(cb.value))
        .map(cb => cb.value);
    
    const costFilters = Array.from(document.querySelectorAll('.filters input[value]:checked'))
        .filter(cb => ['free', 'budget', 'mid', 'premium'].includes(cb.value))
        .map(cb => cb.value);
    
    const roleFilters = Array.from(document.querySelectorAll('.filters input[value]:checked'))
        .filter(cb => ['router', 'query', 'editor', 'research'].includes(cb.value))
        .map(cb => cb.value);
    
    filteredModels = allModels.filter(model => {
        const providerMatch = providerFilters.length === 0 || providerFilters.includes(model.provider);
        
        const totalCost = model.pricing_prompt_per_m_tokens + model.pricing_completion_per_m_tokens;
        let costMatch = false;
        if (costFilters.includes('free') && totalCost === 0) costMatch = true;
        if (costFilters.includes('budget') && totalCost > 0 && totalCost < 1) costMatch = true;
        if (costFilters.includes('mid') && totalCost >= 1 && totalCost <= 5) costMatch = true;
        if (costFilters.includes('premium') && totalCost > 5) costMatch = true;
        if (costFilters.length === 0) costMatch = true;
        
        const roleMatch = roleFilters.length === 0 || 
            model.recommended_for.some(role => roleFilters.includes(role));
        
        return providerMatch && costMatch && roleMatch;
    });
    
    const sortBy = document.getElementById('sort-select').value;
    filteredModels.sort((a, b) => {
        switch (sortBy) {
            case 'cost-asc':
                return (a.pricing_prompt_per_m_tokens + a.pricing_completion_per_m_tokens) - 
                       (b.pricing_prompt_per_m_tokens + b.pricing_completion_per_m_tokens);
            case 'cost-desc':
                return (b.pricing_prompt_per_m_tokens + b.pricing_completion_per_m_tokens) - 
                       (a.pricing_prompt_per_m_tokens + a.pricing_completion_per_m_tokens);
            case 'speed-desc':
                return b.speed_tokens_per_sec - a.speed_tokens_per_sec;
            case 'humaneval-desc':
                return (b.benchmarks?.humaneval || 0) - (a.benchmarks?.humaneval || 0);
            case 'context-desc':
                return b.context_window - a.context_window;
            default:
                return 0;
        }
    });
    
    renderModels();
}

function renderModels() {
    const grid = document.getElementById('model-grid');
    grid.innerHTML = '';
    
    if (filteredModels.length === 0) {
        grid.innerHTML = '<p style="grid-column: 1/-1; text-align: center; padding: 48px;">No models match your filters.</p>';
        return;
    }
    
    filteredModels.forEach(model => {
        const card = document.createElement('div');
        card.className = 'model-card';
        if (selectedModels.some(m => m.id === model.id)) {
            card.classList.add('selected');
        }
        
        const totalCost = model.pricing_prompt_per_m_tokens + model.pricing_completion_per_m_tokens;
        const costDisplay = totalCost === 0 ? 'Free' : `$${totalCost.toFixed(2)}/M`;
        
        card.innerHTML = `
            <div class="model-card-header">
                <div class="model-name">${model.name}</div>
                <div class="model-provider">${model.provider}</div>
            </div>
            
            <div class="model-stats">
                <div class="stat">
                    <div class="stat-label">Cost</div>
                    <div class="stat-value">${costDisplay}</div>
                </div>
                <div class="stat">
                    <div class="stat-label">Speed</div>
                    <div class="stat-value">${model.speed_tokens_per_sec} t/s</div>
                </div>
                <div class="stat">
                    <div class="stat-label">Context</div>
                    <div class="stat-value">${formatContext(model.context_window)}</div>
                </div>
                <div class="stat">
                    <div class="stat-label">HumanEval</div>
                    <div class="stat-value">${model.benchmarks?.humaneval?.toFixed(1) || 'N/A'}</div>
                </div>
            </div>
            
            ${model.benchmarks ? `
            <div class="benchmarks">
                <div class="benchmark">
                    <span class="benchmark-name">SWE-Bench</span>
                    <span class="benchmark-value">${model.benchmarks.swe_bench?.toFixed(1)}%</span>
                </div>
                <div class="benchmark">
                    <span class="benchmark-name">LiveCode</span>
                    <span class="benchmark-value">${model.benchmarks.livecode?.toFixed(1)}%</span>
                </div>
            </div>
            ` : ''}
            
            <div class="model-tags">
                ${model.tags.map(tag => `<span class="tag ${tag}">${tag}</span>`).join('')}
            </div>
        `;
        
        card.addEventListener('click', () => toggleModelSelection(model));
        grid.appendChild(card);
    });
}

function toggleModelSelection(model) {
    const index = selectedModels.findIndex(m => m.id === model.id);
    
    if (index > -1) {
        selectedModels.splice(index, 1);
    } else {
        if (selectedModels.length >= MAX_COMPARISON) {
            alert(`You can compare up to ${MAX_COMPARISON} models at once.`);
            return;
        }
        selectedModels.push(model);
    }
    
    renderModels();
    renderComparison();
}

function renderComparison() {
    const section = document.getElementById('comparison-section');
    const table = document.getElementById('comparison-table');
    
    if (selectedModels.length === 0) {
        section.style.display = 'none';
        return;
    }
    
    section.style.display = 'block';
    
    const metrics = [
        { label: 'Provider', key: 'provider' },
        { label: 'Parameters', key: 'parameters' },
        { label: 'Context Window', key: 'context_window', format: formatContext },
        { label: 'Prompt Cost', key: 'pricing_prompt_per_m_tokens', format: v => `$${v.toFixed(3)}/M`, compare: true },
        { label: 'Completion Cost', key: 'pricing_completion_per_m_tokens', format: v => `$${v.toFixed(3)}/M`, compare: true },
        { label: 'Speed (t/s)', key: 'speed_tokens_per_sec', format: v => v.toString(), compare: true, higher: true },
        { label: 'HumanEval', key: 'benchmarks.humaneval', format: v => `${v.toFixed(1)}%`, compare: true, higher: true },
        { label: 'SWE-Bench', key: 'benchmarks.swe_bench', format: v => `${v.toFixed(1)}%`, compare: true, higher: true },
        { label: 'LiveCode', key: 'benchmarks.livecode', format: v => `${v.toFixed(1)}%`, compare: true, higher: true },
        { label: 'AIME', key: 'benchmarks.aime', format: v => `${v.toFixed(1)}%`, compare: true, higher: true },
        { label: 'Tool Support', key: 'supports_tools', format: v => v ? 'Yes' : 'No' },
        { label: 'JSON Mode', key: 'supports_json', format: v => v ? 'Yes' : 'No' },
    ];
    
    let html = '<table><thead><tr><th>Metric</th>';
    selectedModels.forEach(model => {
        html += `<th>${model.name}</th>`;
    });
    html += '</tr></thead><tbody>';
    
    metrics.forEach(metric => {
        html += `<tr><td>${metric.label}</td>`;
        
        const values = selectedModels.map(model => {
            const keys = metric.key.split('.');
            let value = model;
            for (const key of keys) {
                value = value?.[key];
            }
            return value;
        });
        
        selectedModels.forEach((model, idx) => {
            let value = values[idx];
            let className = '';
            
            if (metric.compare && value != null) {
                const numericValues = values.filter(v => v != null);
                if (metric.higher) {
                    if (value === Math.max(...numericValues)) className = 'best-value';
                    if (value === Math.min(...numericValues)) className = 'worst-value';
                } else {
                    if (value === Math.min(...numericValues)) className = 'best-value';
                    if (value === Math.max(...numericValues)) className = 'worst-value';
                }
            }
            
            const displayValue = value != null && metric.format ? metric.format(value) : (value ?? 'N/A');
            html += `<td class="${className}">${displayValue}</td>`;
        });
        
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    table.innerHTML = html;
}

function calculateCosts() {
    const requestsPerDay = parseInt(document.getElementById('requests-per-day').value) || 0;
    const tokensPerRequest = parseInt(document.getElementById('tokens-per-request').value) || 0;
    
    const totalTokensPerDay = requestsPerDay * tokensPerRequest;
    const totalTokensPerMonth = totalTokensPerDay * 30;
    
    const agentDistribution = {
        router: 0.05,
        query: 0.60,
        editor: 0.30,
        research: 0.05
    };
    
    const results = document.getElementById('calculator-results');
    results.innerHTML = '';
    
    const topModels = [...allModels]
        .filter(m => m.pricing_prompt_per_m_tokens + m.pricing_completion_per_m_tokens > 0)
        .sort((a, b) => {
            const costA = (a.pricing_prompt_per_m_tokens + a.pricing_completion_per_m_tokens) / 2;
            const costB = (b.pricing_prompt_per_m_tokens + b.pricing_completion_per_m_tokens) / 2;
            return costA - costB;
        })
        .slice(0, 4);
    
    topModels.forEach(model => {
        const avgCost = (model.pricing_prompt_per_m_tokens + model.pricing_completion_per_m_tokens) / 2;
        const monthlyCost = (totalTokensPerMonth / 1_000_000) * avgCost;
        
        const resultCard = document.createElement('div');
        resultCard.className = 'cost-result';
        resultCard.innerHTML = `
            <div class="cost-result-model">${model.name}</div>
            <div class="cost-result-value">$${monthlyCost.toFixed(2)}<span>/mo</span></div>
        `;
        results.appendChild(resultCard);
    });
    
    if (totalTokensPerDay === 0) {
        results.innerHTML = '<p style="grid-column: 1/-1; text-align: center; color: var(--color-text-secondary);">Enter request volume to see estimates</p>';
    }
}

function formatContext(tokens) {
    if (tokens >= 1000000) {
        return `${(tokens / 1000000).toFixed(1)}M`;
    } else if (tokens >= 1000) {
        return `${(tokens / 1000).toFixed(0)}K`;
    }
    return tokens.toString();
}

init();
