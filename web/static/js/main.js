const socket = new WebSocket("ws://" + location.host + "/ws");

socket.onmessage = function(event) {
    const state = JSON.parse(event.data);
    updateUI(state);
};

function updateUI(state) {
    const registersView = document.getElementById("registers-view");
    let regHTML = "";
    for (let i = 0; i < state.registers.length; i++) {
        regHTML += `R${i}: ${state.registers[i]}\n`;
    }
    registersView.textContent = regHTML;

    const memoryView = document.getElementById("memory-view");
    memoryView.textContent = state.memory.join(" ");

    document.getElementById("pc-view").textContent = state.pc;
    document.getElementById("running-view").textContent = state.running;
}

