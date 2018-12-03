import React, { Component } from 'react';
import TextField from "@material-ui/core/TextField";
import { withStyles } from "@material-ui/core/styles";
import PropTypes from "prop-types";
import classNames from "classnames";

const styles = theme => ({
	container: {
		display: "flex",
		flexWrap: "wrap"
	},
	textField: {
		marginLeft: theme.spacing.unit,
		marginRight: theme.spacing.unit
	},
	dense: {
		marginTop: 16
	},
	menu: {
		width: 200
	}
});

class TextAreaComponent extends Component {
	constructor() {
		super()
		this.cursorPos = 0
	}

	componentWillUpdate(prevProps) {
		this.cursorPos = this.ref.selectionStart
	}

	componentDidUpdate(prevProps) {
		if (!this.props.handleCursor) {
			return
		}
		if (this.cursorPos < this.props.value.length) {
			const prevVal = prevProps.value
			const newVal = this.props.value
			const charLengthDiff = newVal.length - prevVal.length
			const prevLeftOfStr = prevVal.substring(0, this.cursorPos)
			const newLeftOfStr = newVal.substring(0, this.cursorPos)
			if (prevLeftOfStr !== newLeftOfStr) {
				if (charLengthDiff > 0) {
					//inserted
					console.log("Character inserted: " + charLengthDiff)
					console.log("Old cursor pos: " + this.cursorPos)
					this.cursorPos = this.cursorPos + charLengthDiff
					console.log("New cursor pos: " + this.cursorPos)
				} else if (charLengthDiff < 0) {
					console.log("Character deleted: " + charLengthDiff)
					console.log("Old cursor pos: " + this.cursorPos)
					this.cursorPos =  this.cursorPos - charLengthDiff
					console.log("New cursor pos: " + this.cursorPos)
				}
			}
		}
		var node = this.ref
		node.selectionStart = node.selectionEnd = this.cursorPos
	}

	render() {
		const { classes } = this.props;

		return (
			<form className={classes.container} noValidate autoComplete="off">
			<TextField
			inputRef={ir => this.ref = ir}
			id="outlined-multiline-static"
			label={this.props.title}
			multiline
			rows="25"
			fullWidth={true}
			onChange={this.props.onChange}
			value={this.props.value}
			className={classes.textField}
			margin="normal"
			variant="outlined"
			/>
			</form>
		)
	}
}

TextAreaComponent.propTypes = {
	classes: PropTypes.object.isRequired
};

export default withStyles(styles)(TextAreaComponent);
