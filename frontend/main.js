import { Events } from "@wailsio/runtime";
import { Service as KeyValueStore } from "./bindings/github.com/wailsapp/wails/v3/pkg/services/kvstore";

const resultElement = document.getElementById('result');
const timeElement = document.getElementById('time');

const form = document.getElementById("setup")

KeyValueStore.Get("ha-address").then((value) => {
    if (typeof value !== 'string') value = ''
    form.querySelector('[name="address"]').value = value

    try {
        const url = new URL(value)
        document.querySelector('#connection .host').innerText = `${url.hostname}:${url.port}`
    } catch {
        // don't care
    }
})

KeyValueStore.Get("ha-token").then((value) => {
    if (typeof value !== 'string') value = ''
    form.querySelector('[name="token"]').value = value

    if (value.length > 0) {
        Events.Emit(new Events.WailsEvent("get-status"))
    }
})

form.addEventListener('submit', (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);

    Promise.all([
        KeyValueStore.Set("ha-address", formData.get("address")),
        KeyValueStore.Set("ha-token", formData.get("token"))
    ]).then((res) => {
        console.log('kv: saved values')
        e.submitter.innerText = "Connecting..."
        Events.Emit(new Events.WailsEvent("saved-preferences"))
    }).catch((e) => {
        console.error('kv: failed to save values')
        // handle save failure
    })
})

function setDisplayMode(setupCompleted = true) {
    if (setupCompleted) {
        document.getElementById("active-connection").style = "display: none"
        document.getElementById("setup-mode").style = "display: block"
    } else {
        document.getElementById("active-connection").style = "display: block"
        document.getElementById("setup-mode").style = "display: none"
    }
}

Events.On("ha-status", (event) => {
    document.querySelector('#setup button').innerText = "Connect"
    const statusCode = event.data[0] // why is it an array?

    if (statusCode === 200) {
        resultElement.classList.remove(["error"])

        setDisplayMode(false)
    } else {
        resultElement.classList.add("error")
        resultElement.innerText = `Connection failed! Status code ${statusCode}`

        setDisplayMode(true)
    }

    document.body.setAttribute("data-ha-status", statusCode)
})

Events.On('time', (time) => {
    timeElement.innerText = time.data;
});

document.getElementById("modify-connection").addEventListener("click", (e) => {
    e.preventDefault()
    setDisplayMode(true)
})
