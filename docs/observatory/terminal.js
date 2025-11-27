export function initTerminal() {
    const terminalContent = document.getElementById('terminal-output');
    let charIndex = 0;

    function typeCommand(command) {
        // Clear previous content but keep the prompt line structure if needed, 
        // or just append to the last prompt.
        // For simplicity, let's assume we are typing into the last .command span or creating a new line.

        // Reset state
        charIndex = 0;

        // Create new prompt line
        const line = document.createElement('div');
        line.className = 'terminal-line';
        line.innerHTML = '<span class="prompt">➜</span> <span class="command"></span><span class="cursor">█</span>';
        terminalContent.appendChild(line);

        const commandSpan = line.querySelector('.command');
        const cursorSpan = line.querySelector('.cursor');

        function typeChar() {
            if (charIndex < command.length) {
                commandSpan.textContent += command.charAt(charIndex);
                charIndex++;
                setTimeout(typeChar, 50 + Math.random() * 50);
            } else {
                // Command finished typing
                setTimeout(() => executeCommand(cursorSpan), 500);
            }
        }

        typeChar();
    }

    function executeCommand(cursorSpan) {
        // Remove cursor
        if (cursorSpan) cursorSpan.remove();

        // Add newline
        addToTerminal('<br>');

        // Show "Thinking..."
        const thinkingId = 'thinking-' + Date.now();
        addToTerminal(`<span id="${thinkingId}" class="text-blue-400">Thinking... ⠋</span>`);

        let dots = 0;
        const thinkingInterval = setInterval(() => {
            const el = document.getElementById(thinkingId);
            if (el) {
                const spinner = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
                el.innerHTML = `Thinking... ${spinner[dots % spinner.length]}`;
                dots++;
            }
        }, 100);

        // Simulate processing delay then show logs
        setTimeout(() => {
            clearInterval(thinkingInterval);
            const el = document.getElementById(thinkingId);
            if (el) el.remove();

            streamLogs();
        }, 1500);
    }

    function streamLogs() {
        const logs = [
            { text: '[Analyzer] Reading project map...', delay: 200, color: 'text-blue-400' },
            { text: '[Analyzer] Identified 42 relevant files', delay: 400, color: 'text-blue-400' },
            { text: '[Planner] Creating implementation plan...', delay: 600, color: 'text-purple-400' },
            { text: '[Planner] Plan approved: 3 phases', delay: 800, color: 'text-purple-400' },
            { text: '[Editor] Applying changes to auth/handler.go...', delay: 1000, color: 'text-green-400' },
        ];

        let totalDelay = 0;
        logs.forEach((log, index) => {
            totalDelay += log.delay;
            setTimeout(() => {
                addToTerminal(`<span class="${log.color}">${log.text}</span>`);

                // Trigger visualization after the first few logs
                if (index === 2) {
                    console.log('Terminal dispatching terminal-command-complete event');
                    document.dispatchEvent(new CustomEvent('terminal-command-complete'));
                }
            }, totalDelay);
        });
    }

    function addToTerminal(html) {
        const div = document.createElement('div');
        div.innerHTML = html;
        terminalContent.appendChild(div);
        terminalContent.scrollTop = terminalContent.scrollHeight;
    }

    return {
        typeCommand,
        log: (message, type = 'info') => {
            let color = 'text-gray-400';
            if (type === 'success') color = 'text-green-400';
            if (type === 'error') color = 'text-red-400';
            if (type === 'warning') color = 'text-yellow-400';
            if (type === 'system') color = 'text-blue-400';

            addToTerminal(`<span class="${color}">[${new Date().toLocaleTimeString()}] ${message}</span>`);
        }
    };
}
