/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
