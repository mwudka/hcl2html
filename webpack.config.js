const path = require('path');
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');


module.exports = {
	mode: process.env.NODE_ENV,
	entry: './src/index.js',
	output: {
		path: path.resolve(__dirname, 'dist'),
		filename: '[name].bundle.js',
		publicPath: ''
	},
	module: {
		rules: [
			{
				test: /\.css$/,
				use: [MiniCssExtractPlugin.loader, 'css-loader']
			},
			{
				test: /\.ttf$/,
				use: ['file-loader']
			},
			{
				test: /\.wasm$/,
				use: ['wasm-loader']
			}
		]
	},
	plugins: [
		new MonacoWebpackPlugin({
			languages: ['hcl', 'json']
		}),
		new HtmlWebpackPlugin(),
		new MiniCssExtractPlugin(),
		new CopyWebpackPlugin({
			patterns: [
				{from: 'public/hcl_wasm.wasm', to: './public/hcl_wasm.wasm'}
			]
		})
	]
};
