package html

const indexTemplate = `
<html>
	<head>
	    {{ css }}
		<title>All Frugal Declarations</title>
	</head>
	<body>
		<div class="container-fluid">
			<h1>All Frugal Declarations</h1>
			<table class="table-bordered table-striped table-condensed">
				<tr>
					<th>Module</th>
					<th>Services</th>
					<th>Scopes</th>
					<th>Data Types</th>
					<th>Constants</th>
				</tr>
				{{ range $module := . }}
				<tr>
					<td><a href="{{ $module.Name }}.html">{{ $module.Name }}</a></td>
					<td>
					{{ range $service := $module.Services }}
						<a href="{{ $module.Name }}.html#svc_{{ $service.Name }}">{{ $service.Name }}</a><br />
						<ul>
						{{ range $service.Methods }}
							<li><a href="{{ $module.Name }}.html#fn_{{ $service.Name }}_{{ .Name }}">{{ .Name }}</a></li>
						{{ end }}
						</ul>
					{{ end }}
					</td>
					<td>
					{{ range $scope := $module.Scopes }}
					    <a href="{{ $module.Name }}.html#scp_{{ $scope.Name }}">{{ $scope.Name }}</a><br />
						<ul>
						{{ range $scope.Operations }}
							<li><a href="{{ $module.Name }}.html#fn_{{ $scope.Name }}_{{ .Name }}">{{ .Name }}</a></li>
						{{ end }}
						</ul>
					{{ end }}
					</td>
					<td>
					{{ range $typedef := $module.Typedefs }}
						<a href="{{ $module.Name }}.html#typedef_{{ $typedef.Name }}">{{ $typedef.Name }}</a><br />
					{{ end }}
					{{ range $enum := $module.Enums }}
						<a href="{{ $module.Name }}.html#enum_{{ $enum.Name }}">{{ $enum.Name }}</a><br />
					{{ end }}
					{{ range $struct := $module.DataStructures }}
						<a href="{{ $module.Name }}.html#struct_{{ $struct.Name }}">{{ $struct.Name }}</a><br />
					{{ end }}
					</td>
					<td>
					{{ range $const := $module.Constants }}
					    <code><a href="{{ $module.Name }}.html#const_{{ $const.Name }}">{{ $const.Name }}</a></code><br />
					{{ end }}
					</td>
				</tr>
				{{ end }}
			</table>
		</div>
	</body>
</html>
`
