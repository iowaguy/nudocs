import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import TextAreaComponent from './TextAreaComp.js'
import socketIOClient from "socket.io-client";

const EMIT_TO_SERVER = "client-event"
const RECEIVE_FROM_SERVER = "server-event"



class App extends Component {
	constructor() {
		super()
		this.state = {value: "",message: "Welcome to NuDocs! Start typing to edit!", serverChange: false}
		const endpoint = "http://127.0.0.1:8080"
		this.socket = socketIOClient(endpoint);
	}
	componentDidMount() {
		var self = this
		this.socket.on(RECEIVE_FROM_SERVER, msg=> {
			const arr = msg.split(":")
			const docLength = arr[0]
			var data = ""
			for (var i = 1; i < arr.length; i++) {
				var colonChar = ""
				if (data !== "") {
					colonChar = ":"
				}
				data = data + colonChar + arr[i]
			}
			if (self.state.value !== data) {
				self.setState({...this.state, value: data, serverChange: true})
			}
		});
	}

	handleOnChange = (e) => {
		this.setState({...this.state, value: e.target.value, serverChange: false})
		const prevVal = this.state.value
		const newVal = e.target.value
		const cursorPos = e.target.selectionEnd
		const charLengthDiff = prevVal.length- newVal.length
		const charLengthDiffAbs = Math.abs(charLengthDiff)
		//Take this off to support multiple chars at a time
		if (charLengthDiffAbs > 1) {
			this.setState({...this.state, message: "MultiCharacter operations are not allowed."})
			return
		}
		var emitString = ""
		if (charLengthDiff > 0) {
			//delete operation
			const deletedChars = prevVal.substring(cursorPos, cursorPos + charLengthDiffAbs)
			emitString = "d"+deletedChars+cursorPos+"\n";
			console.log(emitString)
			this.socket.emit(EMIT_TO_SERVER, emitString)
		} else if (charLengthDiff < 0) {
			//insert operation
			const insertedChars = newVal.substring(cursorPos - charLengthDiffAbs, cursorPos)
			emitString = "i"+insertedChars+(cursorPos-charLengthDiffAbs)+"\n"
			console.log(emitString)
			this.socket.emit(EMIT_TO_SERVER, emitString)
		}
	}
	render() {
		return (
			<div className="App" style={{ margin: "50px" }}>
			<TextAreaComponent
			title={this.state.message}
			value={this.state.value}
			onChange={this.handleOnChange}
			handleCursor={this.state.serverChange}
			/>
			</div>
		);
	}
}

export default App;
