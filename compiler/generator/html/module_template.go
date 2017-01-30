package html

const moduleTemplate = `
<html>
	<head>
	    {{ css }}
		<title>Frugal Module: {{ .Name }}</title>
	</head>
	<body>
		<div class="container-fluid">
			<h1>Frugal Module: {{ .Name }}</h1>
			<table class="table-bordered table-striped table-condensed">
				<tr>
					<th>Module</th>
					<th>Services</th>
					<th>Scopes</th>
					<th>Data Types</th>
					<th>Constants</th>
				</tr>
				<tr>
					<td>{{ .Name }}</td>
					<td>
					{{ range $service := .Services }}
						<a href="#svc_{{ $service.Name }}">{{ $service.Name }}</a><br />
						<ul>
						{{ range $service.Methods }}
							<li><a href="#fn_{{ $service.Name }}_{{ .Name }}">{{ .Name }}</a></li>
						{{ end }}
						</ul>
					{{ end }}
					</td>
					<td>
					{{ range $scope := .Scopes }}
					    <a href="#scp_{{ $scope.Name }}">{{ $scope.Name }}</a><br />
						<ul>
						{{ range $scope.Operations }}
							<li><a href="#fn_{{ $scope.Name }}_{{ .Name }}">{{ .Name }}</a></li>
						{{ end }}
						</ul>
					{{ end }}
					</td>
					<td>
					{{ range $typedef := .Typedefs }}
						<a href="#typedef_{{ $typedef.Name }}">{{ $typedef.Name }}</a><br />
					{{ end }}
					{{ range $enum := .Enums }}
						<a href="#enum_{{ $enum.Name }}">{{ $enum.Name }}</a><br />
					{{ end }}
					{{ range $struct := .DataStructures }}
						<a href="#struct_{{ $struct.Name }}">{{ $struct.Name }}</a><br />
					{{ end }}
					</td>
					<td>
					{{ range $const := .Constants }}
					    <code><a href="#const_{{ $const.Name }}">{{ $const.Name }}</a></code><br />
					{{ end }}
					</td>
				</tr>
			</table>
			{{ if .Constants }}
			<hr />
			<h2 id="constants">Constants</h2>
			<table class="table-bordered table-striped table-condensed">
				<tr>
					<th>Constant</th>
					<th>Type</th>
					<th>Value</th>
				</tr>
				{{ range .Constants }}
				<tr id="const_{{ .Name }}">
					<td><code>{{ .Name }}</code></td>
					<td><code>{{ .Type | displayType }}</code></td>
					<td><code>{{ .Value | formatValue }}</code></td>
				</tr>
				{{ if .Comment }}
				<tr>
					<td colspan="3">
						<blockquote>
							{{ range .Comment }}
							{{ . }}<br />
							{{ end }}
						</blockquote>
					</td>
				</tr>
				{{ end }}
				{{ end }}
			</table>
			{{ end }}
			{{ if .Enums }}
			<hr />
			<h2 id="enumerations">Enumerations</h2>
			{{ range .Enums }}
				<div class="definition">
					<h3 id="enum_{{ .Name }}">Enumeration: {{ .Name }}</h3>
					{{ if .Comment }}
						<blockquote>
							{{ range .Comment }}
							{{ . }}<br />
							{{ end }}
						</blockquote>
					{{ end }}
					<table class="table-bordered table-striped table-condensed">
					{{ range .Values }}
						<tr>
							<td><code>{{ .Name }}</code></td>
							<td><code>{{ .Value }}</code></td>
							<td>
							{{ range .Comment }}
								{{ . }}<br />
							{{ end }}
							</td>
						</tr>
					{{ end }}
					</table>
				</div>
			{{ end }}
			{{ end }}
			{{ if .Typedefs }}
			<hr />
			<h2 id="typedefs">Type Declarations</h2>
			{{ range .Typedefs }}
				<div class="definition">
					<h3 id="typedef_{{ .Name }}">Typedef: {{ .Name }}</h3>
					<p>
						<strong>Base type:</strong>&nbsp;
						<code>{{ .Type | displayType }}</code>
					</p>
					{{ if .Comment }}
					<blockquote>
						{{ range .Comment }}
						{{ . }}<br />
						{{ end }}
					</blockquote>
					{{ end }}
				</div>
			{{ end }}
			{{ end }}
			{{ if .DataStructures }}
			<hr />
			<h2 id="structs">Data Structures</h2>
			{{ range .DataStructures }}
			<div class="definition">
				<h3 id="struct_{{ .Name }}">{{ .Type.String | capitalize }}: {{ .Name }}</h3>
				<table class="table-bordered table-striped table-condensed">
					<tr>
						<th>Key</th>
						<th>Field</th>
						<th>Type</th>
						<th>Description</th>
						<th>Requiredness</th>
						<th>Default Value</th>
					</tr>
					{{ range .Fields }}
					<tr>
						<td>{{ .ID }}</td>
						<td>{{ .Name }}</td>
						<td><code>{{ .Type | displayType }}</code></td>
						<td>{{ range .Comment }}{{ . }}<br />{{ end }}</td>
						<td>{{ .Modifier.String | lowercase }}</td>
						<td>{{ if .Default }}<code>{{ .Default | formatValue }}</code>{{ end }}</td>
					</tr>
					{{ end }}
				</table>
				<br />
				{{ if .Comment }}
				<blockquote>
					{{ range .Comment }}
					{{ . }}<br />
					{{ end }}
				</blockquote>
				{{ end }}
			</div>
			{{ end }}
			{{ end }}
			{{ if .Services }}
			<hr />
			<h2 id="services">Services</h2>
			{{ range $service := .Services }}
			<h3 id="svc_{{ $service.Name }}">Service: {{ $service.Name }}</h3>
			{{ if $service.Extends }}
			<div class="extends">
				<em>extends</em> <code>{{ $service.Extends | displayService }}</code>
			</div>
			{{ end }}
			{{ if $service.Comment }}
			<blockquote>
				{{ range $service.Comment }}
				{{ . }}<br />
				{{ end }}
			</blockquote>
			{{ end }}
			{{ range $service.Methods }}
			<div class="definition">
				<h4 id="fn_{{ $service.Name }}_{{ .Name }}">Function: {{ $service.Name }}.{{ .Name }}</h4>
				<pre><code>{{ . | displayMethod }}</code></pre>
				{{ if .Comment }}
				<blockquote>
					{{ range .Comment }}
					{{ . }}<br />
					{{ end }}
				</blockquote>
				{{ end }}
			</div>
			{{ end }}
			{{ end }}
			{{ end }}
			{{ if .Scopes }}
			<hr />
			<h2 id="scopes">Scopes</h2>
			{{ range $scope := .Scopes }}
			<h3 id="scp_{{ $scope.Name }}">Scope: {{ $scope.Name }}</h3>
			{{ if $scope.Prefix.String }}
			<div class="prefix">
				<em>prefix</em> <code>{{ $scope.Prefix.String }}</code>
			</div>
			{{ end }}
			{{ if $scope.Comment }}
			<blockquote>
				{{ range $scope.Comment }}
				{{ . }}<br />
				{{ end }}
			</blockquote>
			{{ end }}
			{{ range $scope.Operations }}
			<div class="definition">
				<h4 id="fn_{{ $scope.Name }}_{{ .Name }}">Operation: {{ $scope.Name }}.{{ .Name }}</h4>
				<pre><code>[publish|subscribe]{{ .Name }}: {{ .Type | displayType }}</code></pre>
				{{ if .Comment }}
				<blockquote>
					{{ range .Comment }}
					{{ . }}<br />
					{{ end }}
				</blockquote>
				{{ end }}
			</div>
			{{ end }}
			{{ end }}
			{{ end }}
		</div>
	</body>
</html>
`
