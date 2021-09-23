const { PulpSocket, events } = require("pulp_web")


const socket = new PulpSocket("mount", "/socket", { debug: false })