export async function loadScenarios() {
    try {
        const response = await fetch('scenarios.json');
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error loading scenarios:', error);
        return {};
    }
}
