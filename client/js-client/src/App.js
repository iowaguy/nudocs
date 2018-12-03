import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import socketIOClient from "socket.io-client";

const EMIT_TO_SERVER = "client-event"
const RECEIVE_FROM_SERVER = "server-event"

class TextAreaComponent extends Component {
	/*	componentDidUpdate(prevProps) {
		const prevValue = prevProps.value
		var node = this.refs.input
		var oldLength = prevValue.length;
		var oldIdx = node.selectionStart;
		node.value = this.props.value;
		var newIdx = Math.max(0, node.value.length - oldLength + oldIdx);
		node.selectionStart = node.selectionEnd = newIdx;
	}
	*/

	render() {
		return <textarea ref="input" onChange={this.props.onChange} value={this.props.value} style={{ width: "1000px", height: "600px" }} />
	}
}

class App extends Component {
	constructor() {
		super()
		this.state = {value: "",message: "Hello world!"}
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
			const data = arr[1]
			if (self.state.value !== data) {
				self.setState({...this.state, value: data})
			}
		});
	}

	handleOnChange = (e) => {
		this.setState({...this.state, value: e.target.value})
		const prevVal = this.state.value
		const newVal = e.target.value
		const cursorPos = e.target.selectionEnd
		const charLengthDiff = prevVal.length- newVal.length
		const charLengthDiffAbs = Math.abs(charLengthDiff)
		//Take this off to support multiple chars at a time
		if (charLengthDiffAbs > 1) {
			this.setState({...this.state, message: "Only one char insertion or deletion allowed at a time"})
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
			<div className="App">
			<h1>{this.state.message}</h1>
			<TextAreaComponent
			value={this.state.value}
			onChange={this.handleOnChange}
			/>
			</div>
		);
	}
}

export default App;
