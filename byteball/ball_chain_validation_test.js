// 1. compose

var compose_unit = {
	unit: {
		version: '1.0',
		alt: '1',
		messages: [{
			app: 'payment',
			payload_location: 'inline',
			payload_hash: 'nHQakA7qoqv7ZYCqwXP4wrbkv2kJLFexGNp13bGTU0M=',
			payload: {
				outputs: [{
					address: '4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F',
					amount: 1896
				}, {address: 'LBHFYR6CFTC5VGZXMIYJCITQZAOJXSNN', amount: 502}],
				inputs: [{
					unit: 'fML9UFInRrLze9bht/ISm7Hs0vv0amvQbZFfNXKXSz8=',
					message_index: 0,
					output_index: 0
				}, {unit: 'qXncju7HMKeW6FERulsWwK035GPD0U2Z2A8uYWjl7Do=', message_index: 0, output_index: 0}]
			}
		}],
		authors: [{
			address: '4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F',
			authentifiers: {r: 'EmwDHUHYKIfXR7gDdDBEGTNTwcpPzwiEDRHl4dxngspVg0QGQtV7Jp2cou96+wlJrcogttXgBAVb6NvjGm3Otw=='}
		}],
		parent_units: ['GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY='],
		last_ball: 'Wc8J321UtBTi83InoPUmSJVZD+BOEPvUGG0bFg1SUgU=',
		last_ball_unit: '/aI5a4xoVF+zO+goazLfV1pDqPmwl1+padr59ahOTzs=',
		witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
		headers_commission: 344,
		payload_commission: 257,
		unit: 'J9s98BJvW4SgS4ectpa1h3cQ9EJN3gtdv31mmRlM6HU='
	}
}

var compose_request = ["request", {
	"command": "post_joint",
	"params": {
		"unit": {
			"version": "1.0",
			"alt": "1",
			"messages": [{
				"app": "payment",
				"payload_location": "inline",
				"payload_hash": "nHQakA7qoqv7ZYCqwXP4wrbkv2kJLFexGNp13bGTU0M=",
				"payload": {
					"outputs": [{
						"address": "4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F",
						"amount": 1896
					}, {"address": "LBHFYR6CFTC5VGZXMIYJCITQZAOJXSNN", "amount": 502}],
					"inputs": [{
						"unit": "fML9UFInRrLze9bht/ISm7Hs0vv0amvQbZFfNXKXSz8=",
						"message_index": 0,
						"output_index": 0
					}, {
						"unit": "qXncju7HMKeW6FERulsWwK035GPD0U2Z2A8uYWjl7Do=",
						"message_index": 0,
						"output_index": 0
					}]
				}
			}],
			"authors": [{
				"address": "4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F",
				"authentifiers": {"r": "EmwDHUHYKIfXR7gDdDBEGTNTwcpPzwiEDRHl4dxngspVg0QGQtV7Jp2cou96+wlJrcogttXgBAVb6NvjGm3Otw=="}
			}],
			"parent_units": ["GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY="],
			"last_ball": "Wc8J321UtBTi83InoPUmSJVZD+BOEPvUGG0bFg1SUgU=",
			"last_ball_unit": "/aI5a4xoVF+zO+goazLfV1pDqPmwl1+padr59ahOTzs=",
			"witness_list_unit": "oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=",
			"headers_commission": 344,
			"payload_commission": 257,
			"unit": "J9s98BJvW4SgS4ectpa1h3cQ9EJN3gtdv31mmRlM6HU=",
			"timestamp": 1512444274
		}
	},
	"tag": "UcvfEW5bgQL/Gkj7eh1vPbY7Pl65B/Wnyx5TrSo2AHQ="
}]

// 2. request get_history

var get_history_request = ["request", {
	"command": "light/get_history",
	"params": {
		"witnesses": [
			"BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3",
			"DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS",
			"FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH",
			"GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN",
			"H5EZTQE7ABFH27AUDTQFMZIALANK6RBG",
			"I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT",
			"JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725",
			"JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC",
			"OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC",
			"S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I",
			"TKT4UESIKTTRALRRLWS4SENSTJX6ODCW",
			"UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ"],
		"addresses": ["4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F", "E374NYF2SQPICVCRKQQT6EO6D5DTFU46", "FAEV3RAJNRGXIKANEKJN6ASSPEIRHWBY", "LBHFYR6CFTC5VGZXMIYJCITQZAOJXSNN", "LUA6IGWUNJSVJQMTSEWEHXR4WQHA5DLR", "W6Z6IDLSXAXBDUIIZJMDE7RN4NCLNCQK"],
		"requested_joints": ["J9s98BJvW4SgS4ectpa1h3cQ9EJN3gtdv31mmRlM6HU="],
		"last_stable_mci": 0,
		"known_stable_units": ["fML9UFInRrLze9bht/ISm7Hs0vv0amvQbZFfNXKXSz8=", "omYI75yfQlPTiMgR4lyiUwsUd/PeGEoqgBEkz6XldaY=", "qXncju7HMKeW6FERulsWwK035GPD0U2Z2A8uYWjl7Do=", "rY3LL+S7RMStrD/tQHVf1AJ/Gi1cBKxJPRIYQ7vRnFs="]
	},
	"tag": "5jGuDVS3t6c+WRtY1SPvBMzB44euRhTNR4DnGZusz/Y="
}]


// 3. response of get_history with ball_chain proof

var the_structure_of_response = {
	unstable_mc_joints: [
		{unit: [Object]}
	],
	witness_change_and_definition_joints: [
		{ unit: [Object], ball: 'CxK1luSnAk5+MaGyaE9wl26JdwAkSPFDqWJdYs9gRng='}
	],
	joints: [{unit: [Object]}],
	proofchain_balls: []
}

var unstable_mc_joints =
	[ { unit:
		{ unit: '561i5dRvVKsykRBdB7t8k+l9sEpA4Hxn4JrEYX8HNEw=',
			version: '1.0',
			alt: '1',
			witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
			last_ball_unit: '/R/d7LJUmBBnpK7RwdwloU/Ds3sQ7kofC6dQnS1pbiU=',
			last_ball: 'qV4L6eTX/d/Nuw85DOppt579+uuFi5iRbuI1CS8Wpv0=',
			headers_commission: 344,
			payload_commission: 217,
			main_chain_index: 1495482,
			timestamp: 1512444784,
			parent_units: [ 'm8wWTH5mV6k1BSLRG8Jw7u9RbMsMP3yC+kY4Ftha6lw=' ],
			authors:
				[ { address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS',
					authentifiers: { r: 'IiL2yEebJEBuLxSFsP8jm1BXQiZax7qbGU0HBLahI8VEDq69j8X3vMaAHeLbXGGcz+69a2vJZEEJj7EqpDCFwA==' } } ],
			messages:
				[ { app: 'payment',
					payload_hash: 'OfN3rNDvtUvf6u3ZRy6P10hpWtPbMxmYnRxFiEB6s/w=',
					payload_location: 'inline',
					payload:
						{ inputs:
							[ { unit: 'sU6oC4MWviRdAPQlZAdih8Ext2lTNbWvol8wiZrAVE0=',
								message_index: 0,
								output_index: 0 },
								{ type: 'headers_commission',
									from_main_chain_index: 1495452,
									to_main_chain_index: 1495460 },
								{ type: 'witnessing',
									from_main_chain_index: 604066,
									to_main_chain_index: 604077 } ],
							outputs: [ { address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 237 } ] } } ] } },
		{ unit:
			{ unit: 'm8wWTH5mV6k1BSLRG8Jw7u9RbMsMP3yC+kY4Ftha6lw=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: '/R/d7LJUmBBnpK7RwdwloU/Ds3sQ7kofC6dQnS1pbiU=',
				last_ball: 'qV4L6eTX/d/Nuw85DOppt579+uuFi5iRbuI1CS8Wpv0=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495481,
				timestamp: 1512444773,
				parent_units: [ 'H0zHH4dNzR92yDH7rnTklEjYW6p1ExnEU6yPAFfNbag=' ],
				authors:
					[ { address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ',
						authentifiers: { r: 'gwaoDqldQ6qaTS33FSvPpyNPJ++8sdezqHrUVxStLME5bPid9b/ru2X6zJ32oIp4ZA2YTLZDh3TyBgjtskf67A==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'oXXbiflbUGpNhbiKeXw8GaHh6Ia/yk8fiF4BEnYHhSE=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'YgShortV7HP3sybrm0kylSOlALgFwfweDwB0o9LdDms=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495451,
										to_main_chain_index: 1495459 },
									{ type: 'witnessing',
										from_main_chain_index: 575157,
										to_main_chain_index: 575162 } ],
								outputs: [ { address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 72 } ] } } ] } },
		{ unit:
			{ unit: 'H0zHH4dNzR92yDH7rnTklEjYW6p1ExnEU6yPAFfNbag=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'gjcTCuVKPRsJd6vy9xjFHvTnVG3bUU1OPXgZ7aednJM=',
				last_ball: 'ULxN7k7vBs1LW4pTbHjVXt1Uc++Ox1/nWUkn2dqyur4=',
				headers_commission: 883,
				payload_commission: 943,
				main_chain_index: 1495480,
				timestamp: 1512444768,
				parent_units: [ 'I+PhuS6/Yb3nQb6wFP/H+ZATGa5+RUJvsHHhQXkpLB0=' ],
				earned_headers_commission_recipients:
					[ { address: 'KZBFQRL7ZFPE7Y52ABA6VRNKTKYRN4AM',
						earned_headers_commission_share: 100 } ],
				authors:
					[ { address: 'AK2ERJ3A6YPMLGUZA6CC667AKQ7RSC77',
						authentifiers: { r: 'xcP0ruuZsoCFVt2QRRtTAjP1UujeBIsaSi1NepdN5P5O6RnUmK4TMXwD6GeD7GzMCH5wXWYjuqMCWocjcWLTyA==' } },
						{ address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV',
							authentifiers: { 'r.0.0': 'xcP0ruuZsoCFVt2QRRtTAjP1UujeBIsaSi1NepdN5P5O6RnUmK4TMXwD6GeD7GzMCH5wXWYjuqMCWocjcWLTyA==' },
							definition:
								[ 'or',
									[ [ 'and',
										[ [ 'address', 'AK2ERJ3A6YPMLGUZA6CC667AKQ7RSC77' ],
											[ 'seen',
												{ what: 'output',
													address: 'JKJPJKRJPIMXZL47CRB4VH7JH5DWPHCS',
													asset: 'base',
													amount: 628000000 } ] ] ],
										[ 'and',
											[ [ 'address', 'JKJPJKRJPIMXZL47CRB4VH7JH5DWPHCS' ],
												[ 'not',
													[ 'seen',
														{ what: 'output',
															address: 'JKJPJKRJPIMXZL47CRB4VH7JH5DWPHCS',
															asset: 'base',
															amount: 628000000 } ] ],
												[ 'in data feed',
													[ [ 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT' ],
														'timestamp',
														'>',
														1513012201936 ] ] ] ] ] ] },
						{ address: 'KZBFQRL7ZFPE7Y52ABA6VRNKTKYRN4AM',
							authentifiers: { r: 'rwFVbwREG07b/g3MUjl0L78TF4qMyFMVgQY81B8wqJtnW1LwbdgA4p8uFpvzBgeBRETIAtxfhTJb2ZcDusjS/A==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'KKELqBsj5vU460kTB+gSU7z2eE5mcXSfhavF7VaOJdY=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: '4jsf0oFJsjRdIHeKKKiSvmwpO2ARTGqbh9Yfcs2FaeA=',
									message_index: 0,
									output_index: 0 } ],
								outputs: [ { address: 'KZBFQRL7ZFPE7Y52ABA6VRNKTKYRN4AM', amount: 13444 } ] } },
						{ app: 'payment',
							payload_hash: '7918dhbxeOjsCeIwmsYL96Q3CO7G9l8rHCBhBLW9UJU=',
							payload_location: 'none',
							spend_proofs:
								[ { spend_proof: 'qRoCBbxboYBhjiAnJ8jE0MoYwJEJOQ5wAltHfgMrms8=',
									address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV' } ] },
						{ app: 'payment',
							payload_hash: 'WdRkCaZVd6Hyi3WGFxiYsjRO5kvxEMqgTBr/KUi6z1w=',
							payload_location: 'none',
							spend_proofs:
								[ { spend_proof: '2s7sQo0Vkpa4vVFxObnMcG81At3n0DjTdSJrGqB/gPw=',
									address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV' } ] },
						{ app: 'payment',
							payload_hash: 'nuHLfxk+KbQBhYt/nrrrrldNUkOyaVr4df5dMU1mzZg=',
							payload_location: 'none',
							spend_proofs:
								[ { spend_proof: 'aoJph8hmRoBxPm49qmLG/lWg48tQE9bq5WD9UNxsZxg=',
									address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV' } ] },
						{ app: 'payment',
							payload_hash: 'KPIqUN6w4/pQI5tSSNB/FQNEJy1SyflxkzTpXKmpL5c=',
							payload_location: 'none',
							spend_proofs:
								[ { spend_proof: '9EeFjhgQxhoL5kUvzupS8N1ooaovth9BZgrHOzh6y+k=',
									address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV' } ] },
						{ app: 'payment',
							payload_hash: 'Yd0snSHrxKUugY64f3vXCjTIkdLWQd4MK9lu4/ciSgg=',
							payload_location: 'none',
							spend_proofs:
								[ { spend_proof: 'QRRcKa4b79s+ejAe4mVBWtzK/FkqBDA7Cvwdj2Txo40=',
									address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV' } ] },
						{ app: 'payment',
							payload_hash: '7xIf/Wvv7qqdHmv16xOu+IP5hxMcjwomaG3+I1DqjvQ=',
							payload_location: 'none',
							spend_proofs:
								[ { spend_proof: 'qa/xDmBCYMwSZRO/xpjK37MnQtjNbP63v74/bFF3CM4=',
									address: 'FFSAEIQRXW37ZFBE6RCKMWM3NEMOPLPV' } ] } ] } },
		{ unit:
			{ unit: 'I+PhuS6/Yb3nQb6wFP/H+ZATGa5+RUJvsHHhQXkpLB0=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'YgShortV7HP3sybrm0kylSOlALgFwfweDwB0o9LdDms=',
				last_ball: 'QYi4DR5lrGUq6rrbvThXTvKvVyrExZYycQ2IMdRmRfA=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495479,
				timestamp: 1512444764,
				parent_units: [ 'g2lh9lkZr8bsxTSf8R35az8u2tKlEhVU52HrP6Kbq+g=' ],
				authors:
					[ { address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I',
						authentifiers: { r: 'rM/SqGXhZoqXTyQPoyoQsS3Uo2Bis+BWquVvYgge/RJOduk8f4MJUxVymAFjOiQkSfKNPEusGwflwJU23pyTOQ==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'ttvHvF+7qFnTzSTfAaN8WZg4Pugam9dkGPYQcLwzsXs=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'ooqiJB/fjh4iRjRNZf4BDAEmGdqTPW3fHDOmiry8HHc=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495440,
										to_main_chain_index: 1495458 },
									{ type: 'witnessing',
										from_main_chain_index: 160173,
										to_main_chain_index: 160187 } ],
								outputs: [ { address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 22 } ] } } ] } },
		{ unit:
			{ unit: 'g2lh9lkZr8bsxTSf8R35az8u2tKlEhVU52HrP6Kbq+g=',
				version: '1.0',
				alt: '1',
				witness_list_unit: '6G8v/UST/CeRCLoNuMD6le3N7mQ4Af6KfdzWJBJ1Tzw=',
				last_ball_unit: 'YgShortV7HP3sybrm0kylSOlALgFwfweDwB0o9LdDms=',
				last_ball: 'QYi4DR5lrGUq6rrbvThXTvKvVyrExZYycQ2IMdRmRfA=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495478,
				timestamp: 1512444747,
				parent_units: [ '+2c1kQmbQBYEFLkNu5joj1EAklKhDOPLJrqm37AUjOg=' ],
				authors:
					[ { address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3',
						authentifiers: { r: 'eO5PY9h5eaoK1LSp8Qj2MowgTqwwI+EYCyHynLNnUIAVU42NuJY9NIoa1g5Lad/mSXnGQ3LBmcW6kba6T8x7xw==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'NdOoZQ1NLQK3S6bDWvD5NKIbwSjJZRMMHYOvafYGOJY=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'f1OzYwR21xqWSVjSzxwJdBNo7pFItJT3wgsJiO32F8k=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495449,
										to_main_chain_index: 1495457 },
									{ type: 'witnessing',
										from_main_chain_index: 562898,
										to_main_chain_index: 562899 } ],
								outputs: [ { address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 68 } ] } } ] } },
		{ unit:
			{ unit: '+2c1kQmbQBYEFLkNu5joj1EAklKhDOPLJrqm37AUjOg=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'ooqiJB/fjh4iRjRNZf4BDAEmGdqTPW3fHDOmiry8HHc=',
				last_ball: 'j2UdJ42y6ujzruVyG97m+zXpnPkRmXhgXfvqia4xAss=',
				headers_commission: 344,
				payload_commission: 258,
				main_chain_index: 1495477,
				timestamp: 1512444733,
				parent_units: [ 'ONvoOiqzNDMRnFszkccQEuLnnjoA9tyB6lxOEghHFBg=' ],
				authors:
					[ { address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT',
						authentifiers: { r: 'M7hRKQQ1SPPK6Kfyhb3/bevK4atbf6d0aFM2d7obXE1/Vot2tnj9/O0kyG7cR6ZA3YQLXDfomnwxoNbMSp1yGw==' } } ],
				messages:
					[ { app: 'data_feed',
						payload_hash: 'mvzMWGpTfGf32isIBuI2PJQ28Cn4T+3oYP8LHxHyVvA=',
						payload_location: 'inline',
						payload: { timestamp: 1512444732865 } },
						{ app: 'payment',
							payload_hash: 'nUuzbRQQCIjbBFxmZnvZ3E3QZXpKPFLeb9vaCq01ZDw=',
							payload_location: 'inline',
							payload:
								{ inputs:
									[ { unit: 'l/Hfua7hAbCIdIL83nxsf4xXyi80dMp5NgUtlVHBhjs=',
										message_index: 1,
										output_index: 0 },
										{ type: 'headers_commission',
											from_main_chain_index: 1495448,
											to_main_chain_index: 1495456 } ],
									outputs: [ { address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 14 } ] } } ] } },
		{ unit:
			{ unit: 'ONvoOiqzNDMRnFszkccQEuLnnjoA9tyB6lxOEghHFBg=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'f1OzYwR21xqWSVjSzxwJdBNo7pFItJT3wgsJiO32F8k=',
				last_ball: 'zZeMvbQ/HyId3sB6trmJecXxW+/cfaWRD+nG5UPQDpk=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495476,
				timestamp: 1512444721,
				parent_units: [ 'I4ZX6AJZu0lspo3Z/qx6wPvM3sdQGKcPcaDkZ77m5Lk=' ],
				authors:
					[ { address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG',
						authentifiers: { r: 'lS1ueRMXyitiHMFGF1u0DJmvl404GdqrFEQYat07Tm0vVAheRn7QezbbBH4IH/76KbsCjWkC69viBD9p06bGEw==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'qOzAO0APATgXogw30DaIBAty24SbuZtTo4QZEAuOWls=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'GFaydbXz9YzkTDfFT4Xliul2d8hO9iZl0LWbE3ZSf7w=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495446,
										to_main_chain_index: 1495455 },
									{ type: 'witnessing',
										from_main_chain_index: 538783,
										to_main_chain_index: 538787 } ],
								outputs: [ { address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 12 } ] } } ] } },
		{ unit:
			{ unit: 'I4ZX6AJZu0lspo3Z/qx6wPvM3sdQGKcPcaDkZ77m5Lk=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'l/Hfua7hAbCIdIL83nxsf4xXyi80dMp5NgUtlVHBhjs=',
				last_ball: 'lsC49OVxAU2d/Y7wr80LDPScq/MpNT8EuWLBQhXU514=',
				headers_commission: 344,
				payload_commission: 191,
				main_chain_index: 1495475,
				timestamp: 1512444709,
				parent_units: [ 'hJ8+R1ILJsj5dP24THJCjYcaOqqgzKBNwyvQiIduo34=' ],
				authors:
					[ { address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW',
						authentifiers: { r: 'hlYo1pH+sH7E7QP3865EKLCw2ifWB7+wfiKxKJvMlQwawOfqmu/U7Scl6iJnyUtd6AuBs7/OBjNHbqR/5NU4CQ==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'xTw0s2jeVN6d9zjFFKcch64QQp1fAjQp0cMOrmo4b4E=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'pY7YNLYkwEsCPSkOJReWNecC2mZyTFTM+AdzBqZUAF8=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495436,
										to_main_chain_index: 1495454 } ],
								outputs: [ { address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 179 } ] } } ] } },
		{ unit:
			{ unit: 'hJ8+R1ILJsj5dP24THJCjYcaOqqgzKBNwyvQiIduo34=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'GFaydbXz9YzkTDfFT4Xliul2d8hO9iZl0LWbE3ZSf7w=',
				last_ball: 'afHf32n8vkXhtY7Y/Io7h55MyBakX3eOJ/1mxodHIRM=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495474,
				timestamp: 1512444703,
				parent_units: [ 'pD7XpwE3eSOyBqWnL5uttzqRD97JxG3KpShs5GoxW30=' ],
				authors:
					[ { address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725',
						authentifiers: { r: '0O7AlJPsB1jsCoeJKOy1nvZG2l8K9Ix9ax+Q3Q1Wm3gHQi/M8txl8Alk76i/ce8QgFPp9P6MIYwZFbDiEgu/yA==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'hYPL3GyJlOTPwKjLqrWkVMPApJ+2GTY/Jjf10bALPfU=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'Yd0uW0PTzw8wz019SqroMETA3wn964/FwdXxGK2HzDQ=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495435,
										to_main_chain_index: 1495453 },
									{ type: 'witnessing',
										from_main_chain_index: 143295,
										to_main_chain_index: 143303 } ],
								outputs: [ { address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 10 } ] } } ] } },
		{ unit:
			{ unit: 'pD7XpwE3eSOyBqWnL5uttzqRD97JxG3KpShs5GoxW30=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'GFaydbXz9YzkTDfFT4Xliul2d8hO9iZl0LWbE3ZSf7w=',
				last_ball: 'afHf32n8vkXhtY7Y/Io7h55MyBakX3eOJ/1mxodHIRM=',
				headers_commission: 344,
				payload_commission: 191,
				main_chain_index: 1495473,
				timestamp: 1512444689,
				parent_units: [ 'BKr8INAy6ubOkZm8Z5h11nTSHB6FV6ekAmlgK5L2znU=' ],
				authors:
					[ { address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC',
						authentifiers: { r: '48nOzPpgRU/wVArLfJEtWmcTvNAMXjhGuIBXIMDpu4ww3QjCVbGURmiiU6ae7xJ44E0qMtko8TtmcQ3GcSv5dg==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'mrWk+gHjpWnuc2kBDonCRvVNTQtLfrbfPNEirnt7Bgw=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'GjW+VD9hIW5n7DlDPJYCWSM7YH7x8mX9E+AmiHJnWB4=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495443,
										to_main_chain_index: 1495452 } ],
								outputs: [ { address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 286 } ] } } ] } },
		{ unit:
			{ unit: 'BKr8INAy6ubOkZm8Z5h11nTSHB6FV6ekAmlgK5L2znU=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'pY7YNLYkwEsCPSkOJReWNecC2mZyTFTM+AdzBqZUAF8=',
				last_ball: '6Vkgi1n5c0vOl4z6PFjV7p3m79Fry0DjRu0s2WLXQf0=',
				headers_commission: 598,
				payload_commission: 257,
				main_chain_index: 1495472,
				timestamp: 1512444686,
				parent_units: [ 'VPke2hisfWWxyMIptV1udOemJ05o2fSztg5n2py4dLI=' ],
				earned_headers_commission_recipients:
					[ { address: 'TTKHI34RMP4GCFTGDU34XAWOGYOK46Z5',
						earned_headers_commission_share: 100 } ],
				authors:
					[ { address: '5FTII3FEH5LT5WVBHUZUPKZRDHPFK4QT',
						authentifiers: { r: 'M/OLE64dwqfcJCPiZ2F/ZfpUKX35XKSjMcHe8FH5qZE3ipA0EuVPk0DI8D9n5UQh7P4KTft8z/hyJyi5HrJfQw==' },
						definition:
							[ 'sig',
								{ pubkey: 'A+1xAZnxNBatfwIlZNgXWXw616bEq4c/POPNfNiv4/U0' } ] },
						{ address: 'J5QGFTZPXGWDM6ETOXKAMRLK6V7POVAQ',
							authentifiers: { r: '8kpsWBVxPXZk+Av96rblg1+p+CEtTwdfGNIR45YcCgY84JwFggbMS9DLf21zD/7iue79+IqguJVccFCfi3Wz5g==' },
							definition:
								[ 'sig',
									{ pubkey: 'AvEsUT4G10nn8GyseEg5mkPQdPAbz2XuTNy2JJW18Vez' } ] } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'wRskX/ObQ3qQnmXlDjNFo13GEFjzJX65jkqRplBindg=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'J2uEly5J2j4lIkDbzYoHdaNelyp2Es+2E+RsU/DTM7c=',
									message_index: 0,
									output_index: 0 },
									{ unit: 'A+oPFUzNJxrx4xMLImpav/8p2zhBt2/tK6Ey9qYywQA=',
										message_index: 0,
										output_index: 1 } ],
								outputs:
									[ { address: 'BKAIEWKS7O3OEBUYEGN6PN2LSWXD6BXM',
										amount: 631470000 },
										{ address: 'TTKHI34RMP4GCFTGDU34XAWOGYOK46Z5',
											amount: 352201998 } ] } } ] } },
		{ unit:
			{ unit: 'VPke2hisfWWxyMIptV1udOemJ05o2fSztg5n2py4dLI=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'Yd0uW0PTzw8wz019SqroMETA3wn964/FwdXxGK2HzDQ=',
				last_ball: 'H4MSM/pfJhJbTvi7tb9udMla1i9ITltTGhYkhVMXfZ4=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495471,
				timestamp: 1512444672,
				parent_units: [ 'WlZwp7fB/GTC9TyhxCWcupEGjCwhNrm53JWrAa56dXI=' ],
				authors:
					[ { address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS',
						authentifiers: { r: '2l8zZaOOA41EKhpWk53i4DidRJyvDRsh0NJnjWAMcAVJ84tyUid8ED/IYfu0eze1uf1sAzfg0u3SEfvehvHmLw==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'eKKuRKCILfI/Wv0vIeUBDnDnBOloacWE9vIOYExjx9w=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'icDMJzNozUcHbO7p+ytO6hGFRl0+pZLw2bY8KnbK58I=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495442,
										to_main_chain_index: 1495451 },
									{ type: 'witnessing',
										from_main_chain_index: 604053,
										to_main_chain_index: 604065 } ],
								outputs: [ { address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 16 } ] } } ] } },
		{ unit:
			{ unit: 'WlZwp7fB/GTC9TyhxCWcupEGjCwhNrm53JWrAa56dXI=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'Yd0uW0PTzw8wz019SqroMETA3wn964/FwdXxGK2HzDQ=',
				last_ball: 'H4MSM/pfJhJbTvi7tb9udMla1i9ITltTGhYkhVMXfZ4=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495470,
				timestamp: 1512444659,
				parent_units: [ 'vJ/cQXbSizcrPICWSc/gl5v6PkpGbZDQpVsxdVkUnts=' ],
				authors:
					[ { address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ',
						authentifiers: { r: 'BAOmWOxIj1g7ATYhCCuoTcjunh7YGYJx6wMtrJRSDx8RxIB6mR6jFBvf/bVQ2zGA9CN//8ESHPoYP9FKYAG3wQ==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'OdJ2Xo/xLsHXaVM6HwT7fYxa5qY1z/2rOzlknyCcERA=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: '0mYoleSLyBpZhRlFbI9UmP04o2at7yAeD9dA+8qgH84=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495441,
										to_main_chain_index: 1495450 },
									{ type: 'witnessing',
										from_main_chain_index: 575153,
										to_main_chain_index: 575156 } ],
								outputs: [ { address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 98 } ] } } ] } },
		{ unit:
			{ unit: 'vJ/cQXbSizcrPICWSc/gl5v6PkpGbZDQpVsxdVkUnts=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'GjW+VD9hIW5n7DlDPJYCWSM7YH7x8mX9E+AmiHJnWB4=',
				last_ball: 'aLoMLcgGmH7V8a7bTdXntSxYgWY8NvSHJ89QiekpjjY=',
				headers_commission: 344,
				payload_commission: 256,
				main_chain_index: 1495469,
				timestamp: 1512444647,
				parent_units: [ '+3T2APZDOfKWYcWkQCDTRigTYasydaskO2jnrDt9JVc=' ],
				authors:
					[ { address: 'MA5VSS2WDH7DE6Q74MZXTU4LUXOYORKQ',
						authentifiers: { r: 'nxQ9OH3p6VUAsV1e3flaYocyF3Jlg4+2pQa5qaqo68EA9PLOXShbWb05qMEzjMr0z6fDx5bgj99lF1LxjYQfig==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: '4oUKJjIR1AFVALKO/W9ZotpyVlWwlog2nras01SbIuM=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'KQs2WSTpNVzFhmT/LAWo6lRJZIzkWqgSlHGMeynY3uY=',
									message_index: 0,
									output_index: 1 } ],
								outputs:
									[ { address: 'MA5VSS2WDH7DE6Q74MZXTU4LUXOYORKQ',
										amount: 35586589 } ] } },
						{ app: 'payment',
							payload_hash: 'dxCN6tEmOpiDfxWeHeoEllUReFQU2a9F0dnebXuxAvM=',
							payload_location: 'none',
							spend_proofs: [ { spend_proof: '/4mEa8msNdOH3vsmlzfq87pmDnGgMmH3rEZs6bRfS64=' } ] } ] } },
		{ unit:
			{ unit: '+3T2APZDOfKWYcWkQCDTRigTYasydaskO2jnrDt9JVc=',
				version: '1.0',
				alt: '1',
				witness_list_unit: '6G8v/UST/CeRCLoNuMD6le3N7mQ4Af6KfdzWJBJ1Tzw=',
				last_ball_unit: 'icDMJzNozUcHbO7p+ytO6hGFRl0+pZLw2bY8KnbK58I=',
				last_ball: '2f1EnVelL4Uli1EhosrL41BNmVJrDT227lcCyb7r2O8=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495468,
				timestamp: 1512444641,
				parent_units: [ 'f72F2q30vlKujJH36WOW7FCG7mLNpuvSSevRC6ZGLIY=' ],
				authors:
					[ { address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3',
						authentifiers: { r: '49AhhnPCPldGWg4C5e1TNgC9DtLnEOtZQc+Q416gRLgII9ZjigrWzbbtrx4ABTF9vHg7gnBrJGbd14568QmPaw==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: '3Y8+FIfu7HMQuJyhI2vGcuWw1LFyhXezTA6BqYSNdC8=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495439,
										to_main_chain_index: 1495448 },
									{ type: 'witnessing',
										from_main_chain_index: 562892,
										to_main_chain_index: 562897 } ],
								outputs: [ { address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 28 } ] } } ] } },
		{ unit:
			{ unit: 'f72F2q30vlKujJH36WOW7FCG7mLNpuvSSevRC6ZGLIY=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY=',
				last_ball: 'bxApGP0z6bSFbVbp/qAbHXsrk/ZuhtdJrre3/tPeFwc=',
				headers_commission: 344,
				payload_commission: 284,
				main_chain_index: 1495467,
				timestamp: 1512444623,
				parent_units: [ 'jB0Ood/01C1HByV1gaJyPSLevIu4IA3vQZG47C/E31c=' ],
				authors:
					[ { address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT',
						authentifiers: { r: 'XQXNF3W//o6MDQVeX4xWu962gGKxoP62Zw7wL8l+EIdDAVZisTEJRstOL4Wu8zfB6/QwjoWdPrMGIEgPRGWy6A==' } } ],
				messages:
					[ { app: 'data_feed',
						payload_hash: 's5zdp+FhhljSCn1QSSBNcdwxqCAE0ovrL7bLA/OEEOg=',
						payload_location: 'inline',
						payload: { timestamp: 1512444622955 } },
						{ app: 'payment',
							payload_hash: 'Phs02hp25RzvaiywllzZuidyucF6VPnTD9NVXj2SLJs=',
							payload_location: 'inline',
							payload:
								{ inputs:
									[ { unit: 'H0QcDkw8s/zIzpD07xmxLAfi7u2viFT01sk6QmlHNzo=',
										message_index: 1,
										output_index: 0 },
										{ type: 'headers_commission',
											from_main_chain_index: 1495438,
											to_main_chain_index: 1495447 },
										{ type: 'witnessing',
											from_main_chain_index: 912102,
											to_main_chain_index: 912102 } ],
									outputs: [ { address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 40 } ] } } ] } },
		{ unit:
			{ unit: 'jB0Ood/01C1HByV1gaJyPSLevIu4IA3vQZG47C/E31c=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY=',
				last_ball: 'bxApGP0z6bSFbVbp/qAbHXsrk/ZuhtdJrre3/tPeFwc=',
				headers_commission: 344,
				payload_commission: 217,
				main_chain_index: 1495466,
				timestamp: 1512444604,
				parent_units:
					[ '6fIzeTgPB6yUhRmeF2iXmZR+ZTju1Fpway91syHzPeA=',
						'7J0W8Q0OiKOVJpNXbvuhhtiyq4pENZMTG3WRm2j05Ig=' ],
				authors:
					[ { address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG',
						authentifiers: { r: '8Ma9MCC6PzSSyVMNaYbn3l+RUgsIIaZ6r+QsLttDZ8cVohcXnWWqYsOz520wqnEZ4xcdPY88IA3Mj83I92T74g==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'dezPXYHvHn3chS0Vm1vrI93r3h1Wfq/vDyogf5zqM4o=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'fCvGL8Ud3M4OYjUMNBGAwgG0KQXdpUz4fkkICttzFCk=',
									message_index: 0,
									output_index: 0 },
									{ type: 'headers_commission',
										from_main_chain_index: 1495437,
										to_main_chain_index: 1495445 },
									{ type: 'witnessing',
										from_main_chain_index: 538770,
										to_main_chain_index: 538782 } ],
								outputs: [ { address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 7 } ] } } ] } },
		{ unit:
			{ unit: '6fIzeTgPB6yUhRmeF2iXmZR+ZTju1Fpway91syHzPeA=',
				version: '1.0',
				alt: '1',
				witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				last_ball_unit: 'GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY=',
				last_ball: 'bxApGP0z6bSFbVbp/qAbHXsrk/ZuhtdJrre3/tPeFwc=',
				headers_commission: 344,
				payload_commission: 197,
				main_chain_index: 1495465,
				timestamp: 1512444593,
				parent_units: [ 'DUJLq5ITUFgsQYQpGP7VQtWwaQet3eQsr9pdaMEGDag=' ],
				authors:
					[ { address: 'QR542JXX7VJ5UJOZDKHTJCXAYWOATID2',
						authentifiers: { r: 'bFImSb8o9X+I5DWEwvGUnnoNCzY/Ie+1OgGy0b6990Yv+D9wtUKnQ3R4GgzcVVMDMplX4G8lXAcK+4v31pF38Q==' } } ],
				messages:
					[ { app: 'payment',
						payload_hash: 'yhhEl4zfD7ukrwjyKLrR184Lwaq3+aS/BK20Ohnpd9Y=',
						payload_location: 'inline',
						payload:
							{ inputs:
								[ { unit: 'QhTTvy3Rwd5FuR+SzmiJMOijWFVieusCXYNwp/ADgEM=',
									message_index: 0,
									output_index: 0 } ],
								outputs:
									[ { address: 'OMNKOLQVVRR7SAVX4DLYLDVP6BQEI245',
										amount: 100000000 },
										{ address: 'QR542JXX7VJ5UJOZDKHTJCXAYWOATID2',
											amount: 270935056033 } ] } } ] } }
											]

var witness_change_and_definition_joints = [
	{
		unit:
			{
				unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
				version: '1.0',
				alt: '1',
				headers_commission: 2520,
				payload_commission: 5143,
				main_chain_index: 0,
				timestamp: 1482617405,
				witnesses:
					['BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3',
						'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS',
						'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH',
						'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN',
						'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG',
						'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT',
						'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725',
						'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC',
						'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC',
						'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I',
						'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW',
						'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ'],
				earned_headers_commission_recipients:
					[{
						address: 'TUOMEGAZPYLZQBJKLEM2BGKYR2Q5SEYS',
						earned_headers_commission_share: 100
					}],
				authors:
					[{
						address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3',
						authentifiers: {r: 'OhYiMWB3Fcy37r3qhzz4iTt1zq5W81EJvI0xEVW9LEtJHPMjDf+at0mld5uWkqWS79mL+uv947Ie9kVY2Cximw=='},
						definition:
							['sig',
								{pubkey: 'A/2sUkxKT6bo5KvTO2b4iprDKdyTlszuRh+O2A1KKn1I'}]
					},
						{
							address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS',
							authentifiers: {r: 'WHh6yLbY7YlaLb/UXPA1t9mTVs7hwxlvxYuQzfDzXm4PmWklVG6UKsJ6av3Er4mNgL40rLU029TN+VEhex9vrw=='},
							definition:
								['sig',
									{pubkey: 'AwTz/u/JbunP3JmxrC3+cPO4ttIYzbYkUvCm7t5Mavib'}]
						},
						{
							address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH',
							authentifiers: {r: 'HGBUPvTCzKT6dqfwXJGm/xRsWQcO+efGIrgIIy/Na9AiLbxUs+wTy1fttma+nuY7sZbfKu0GSjJL/Pgegvrr8g=='},
							definition:
								['sig',
									{pubkey: 'AupKK9N7BQNrZ5mi0dVYjDmED6a+9xFJ+6YTdeeS8yTF'}]
						},
						{
							address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN',
							authentifiers: {r: 'gRAlQOYE0FHNvrQC777vkyuuZmIfkJoE5vmt5HEWLwkN95PXczjuqjy90vVw5fOUWT5jWJzp1TwoPPzkrWhsUQ=='},
							definition:
								['sig',
									{pubkey: 'Aj5MbcDBh++lvWMjiwVWZVEtDRwFZ0bUre28Ut262Dym'}]
						},
						{
							address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG',
							authentifiers: {r: 'YaAFrXvDfIdhvm7UJc5Tx+x5wGnCdHA2BDfJEBPAnPRID+WMtOMsOe0CP+Ef9QHbSRColQ4SkFoTnCx7lAtrCQ=='},
							definition:
								['sig',
									{pubkey: 'Avp4o7tAQn4jFvMB+pfCAyqQ1N+Xnbc/v1rBQlMfieMG'}]
						},
						{
							address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT',
							authentifiers: {r: 'FepHM77ENfueJlBDD91MEufSJc50e/XQ3laW7zBgfAAc26wyVvA/DwkxudzjBtK8xfrSA/mGs3i5YbEw7fHstQ=='},
							definition:
								['sig',
									{pubkey: 'AhrSHfYgsmMWFF+REgoN9j18JxfIxtQNR2xBAnLv8XlR'}]
						},
						{
							address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725',
							authentifiers: {r: '6jbGxqxNaoEkM7s82DnJxHH9NJxmvzzqAM9/76L4NFpvfPqgUhSfWLQNPUTFDYyRYJ0I+XrvwVBgyYs/gvIgaA=='},
							definition:
								['sig',
									{pubkey: 'Aiy5z+jM1ySTZl1Qz1YZJouF7tU6BU++SYc/xe0Rj5OZ'}]
						},
						{
							address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC',
							authentifiers: {r: '0uk8Lm9yH8FmvZYEF0as+3nlm1Y1jx0vz+woW/vZJBpt/OU8rDmSZv1TRCVAlcfLJ2slIYk9BztAVlUvqX0j6A=='},
							definition:
								['sig',
									{pubkey: 'A+ADvohE7iJ3+UHt9DUQK4BOD+vNr1Qhl3kTiqbo9r3R'}]
						},
						{
							address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC',
							authentifiers: {r: 'ZhhrNDkXD9p6v4DGZUeG0A5Kcdf0PJ+3J/Q1snEjbo56SVndzbAQpieTcvPl21vjSUIdI+WUu0vYv/b/8aWmkQ=='},
							definition:
								['sig',
									{pubkey: 'AxIZY4OmlvoL3YjR8n4+4IxXT9YhMmlqyPnhS7B8UM/T'}]
						},
						{
							address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I',
							authentifiers: {r: 'rv+ZRlLvXj5ZPVMuWmirtdffQ4G+REMvT+4IbDQLxAA/81UBGQhNTT2Tj9xXEGQwmNcgiN/Hx+Kpt/UEEOTPxA=='},
							definition:
								['sig',
									{pubkey: 'AlMVfH3AmDm9jGjgg6/aTPoXFRwKaEKbgfzocrfjOp6o'}]
						},
						{
							address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW',
							authentifiers: {r: '7ZnZMUoUA/gTRSK7qgZWz4+HKOeMYPLpJQ5D1ZQyflVL9SQwYjU8kZ6SBznziti838CBq1IGJhF3b2UAM39ehQ=='},
							definition:
								['sig',
									{pubkey: 'A/ow4CFB2YXcY4GPkkcUKquH9/TehAz13YyVWYsAp78J'}]
						},
						{
							address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ',
							authentifiers: {r: 'r/inl+0PEWpD9jIoJzFCvs9bYYbTrpLMY81VsYKHHddWe6pDwH9Qb5eiG+ZoeiZtS+KOth3VBKyku7NL2LEiTQ=='},
							definition:
								['sig',
									{pubkey: 'A3mDugbotUTdc90pnOC9LCkEJPNm3OOTRx3f6SlqpJF+'}]
						}],
				messages:
					[{
						app: 'text',
						payload_hash: '7yUoyu41A83xzX/lIK4psWnMktgsY2AyHSrWC8cDFwk=',
						payload_location: 'inline',
						payload: 'Let there be light!'
					},
						{
							app: 'payment',
							payload_hash: 'gUo+Tj2dJzk84RYCRkD8HbiYOe8g1gl0YCRSHsLAJpE=',
							payload_location: 'inline',
							payload:
								{
									inputs:
										[{
											type: 'issue',
											serial_number: 1,
											amount: 1000000000000000,
											address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3'
										}],
									outputs:
										[{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{address: 'BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3', amount: 1000000},
											{
												address: 'BZUAVP5O4ND6N3PVEUZJOATXFPIKHPDC',
												amount: 10000000000000
											},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'H5EZTQE7ABFH27AUDTQFMZIALANK6RBG', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{address: 'JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC', amount: 1000000},
											{
												address: 'MZ4GUQC7WUKZKKLGAS3H3FSDKLHI7HFO',
												amount: 880000000000000
											},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{address: 'TKT4UESIKTTRALRRLWS4SENSTJX6ODCW', amount: 1000000},
											{
												address: 'TUOMEGAZPYLZQBJKLEM2BGKYR2Q5SEYS',
												amount: 9999869992337
											},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{address: 'UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ', amount: 1000000},
											{
												address: 'UGP3YFIJCOCUGLMJUFKLXLNYNO4S7PT6',
												amount: 100000010000000
											}]
								}
						}]
			},
		ball: 'CxK1luSnAk5+MaGyaE9wl26JdwAkSPFDqWJdYs9gRng='
	}
]


var joints = [{
	unit: {
		unit: 'J9s98BJvW4SgS4ectpa1h3cQ9EJN3gtdv31mmRlM6HU=',
		version: '1.0',
		alt: '1',
		witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=',
		last_ball_unit: '/aI5a4xoVF+zO+goazLfV1pDqPmwl1+padr59ahOTzs=',
		last_ball: 'Wc8J321UtBTi83InoPUmSJVZD+BOEPvUGG0bFg1SUgU=',
		headers_commission: 344,
		payload_commission: 257,
		main_chain_index: 1495450,
		timestamp: 1512444273,
		parent_units: ['GWy5KPFAuMWO/VPwSutnG9CGW/IaKf1bhnU0nUU8bNY='],
		authors:
			[{
				address: '4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F',
				authentifiers: {r: 'EmwDHUHYKIfXR7gDdDBEGTNTwcpPzwiEDRHl4dxngspVg0QGQtV7Jp2cou96+wlJrcogttXgBAVb6NvjGm3Otw=='}
			}],
		messages:
			[{
				app: 'payment',
				payload_hash: 'nHQakA7qoqv7ZYCqwXP4wrbkv2kJLFexGNp13bGTU0M=',
				payload_location: 'inline',
				payload: {
					inputs: [{
						unit: 'fML9UFInRrLze9bht/ISm7Hs0vv0amvQbZFfNXKXSz8=',
						message_index: 0,
						output_index: 0
					}, {
						unit: 'qXncju7HMKeW6FERulsWwK035GPD0U2Z2A8uYWjl7Do=',
						message_index: 0,
						output_index: 0
					}],
					outputs: [{
						address: '4JFF2HIQYHC2S7ZX7ZRYHIJV4P7VGX2F', amount: 1896
					}, {
						address: 'LBHFYR6CFTC5VGZXMIYJCITQZAOJXSNN', amount: 502
					}]
				}
			}]
	}
}]

var proofchain_balls = [
	{
		unit: 'l/Hfua7hAbCIdIL83nxsf4xXyi80dMp5NgUtlVHBhjs=',
		ball: 'lsC49OVxAU2d/Y7wr80LDPScq/MpNT8EuWLBQhXU514=',
		parent_balls: ['afHf32n8vkXhtY7Y/Io7h55MyBakX3eOJ/1mxodHIRM=']
	},
	{
		unit: 'GFaydbXz9YzkTDfFT4Xliul2d8hO9iZl0LWbE3ZSf7w=',
		ball: 'afHf32n8vkXhtY7Y/Io7h55MyBakX3eOJ/1mxodHIRM=',
		parent_balls: ['6Vkgi1n5c0vOl4z6PFjV7p3m79Fry0DjRu0s2WLXQf0=']
	},
	{
		unit: 'pY7YNLYkwEsCPSkOJReWNecC2mZyTFTM+AdzBqZUAF8=',
		ball: '6Vkgi1n5c0vOl4z6PFjV7p3m79Fry0DjRu0s2WLXQf0=',
		parent_balls: ['H4MSM/pfJhJbTvi7tb9udMla1i9ITltTGhYkhVMXfZ4=']
	},
	{
		unit: 'Yd0uW0PTzw8wz019SqroMETA3wn964/FwdXxGK2HzDQ=',
		ball: 'H4MSM/pfJhJbTvi7tb9udMla1i9ITltTGhYkhVMXfZ4=',
		parent_balls: ['aLoMLcgGmH7V8a7bTdXntSxYgWY8NvSHJ89QiekpjjY=']
	},
	{
		unit: 'GjW+VD9hIW5n7DlDPJYCWSM7YH7x8mX9E+AmiHJnWB4=',
		ball: 'aLoMLcgGmH7V8a7bTdXntSxYgWY8NvSHJ89QiekpjjY=',
		parent_balls: ['2f1EnVelL4Uli1EhosrL41BNmVJrDT227lcCyb7r2O8=']
	},
	{
		unit: 'icDMJzNozUcHbO7p+ytO6hGFRl0+pZLw2bY8KnbK58I=',
		ball: '2f1EnVelL4Uli1EhosrL41BNmVJrDT227lcCyb7r2O8=',
		parent_balls: ['HdJERhZUsNqgRXAiFsSuNdOtO7GWMgDv+pgvQ3+Cquo=']
	},
	{
		unit: '0mYoleSLyBpZhRlFbI9UmP04o2at7yAeD9dA+8qgH84=',
		ball: 'HdJERhZUsNqgRXAiFsSuNdOtO7GWMgDv+pgvQ3+Cquo=',
		parent_balls: ['OwHR2mSXgwmcTcagG0Jutm5x5/6fkkSbkyIWX+TEm7E=']
	},
	{
		unit: 'J9s98BJvW4SgS4ectpa1h3cQ9EJN3gtdv31mmRlM6HU=',
		ball: 'OwHR2mSXgwmcTcagG0Jutm5x5/6fkkSbkyIWX+TEm7E=',
		parent_balls: ['bxApGP0z6bSFbVbp/qAbHXsrk/ZuhtdJrre3/tPeFwc='],
		skiplist_balls: ['rSlG2LJ4HABx/nvEgsCqt2CXkZgByzPzvAC0E7uxXS0=']
	}
]



// test validation
var assert = require('assert');
var util= require('util')
//var conf = require('byteballcore/conf.js');
var constants = require('byteballcore/constants.js');
var objectHash = require('./my_hash.js');
var ecdsaSig = require('byteballcore/signature.js');

var objUnit = unstable_mc_joints[0].unit

assert.equal(objectHash.getUnitHash(objUnit),objUnit.unit)
// 561i5dRvVKsykRBdB7t8k+l9sEpA4Hxn4JrEYX8HNEw=
console.log(objectHash.getUnitHash(objUnit)+ " === " +objUnit.unit)

// please Notice : those fields are removed before generate hash.
// delete objNakedUnit.unit;
// delete objNakedUnit.headers_commission;
// delete objNakedUnit.payload_commission;
// delete objNakedUnit.main_chain_index;
// delete objNakedUnit.timestamp;
console.log(objUnit.main_chain_index)
objUnit.main_chain_index=99999
console.log(objUnit.main_chain_index)
assert.equal(objectHash.getUnitHash(objUnit),objUnit.unit)
console.log("hash is still",objUnit.unit)
// it means the unit hash is created by the unit author, instead of anyone else. bacause the unit author don't know the
// mci index

assert.equal(objUnit.parent_units[0],unstable_mc_joints[1].unit.unit)
console.log("parent is "+objUnit.parent_units[0])

var arrWitnesses = [
	"BVVJ2K7ENPZZ3VYZFWQWK7ISPCATFIW3",
	"DJMMI5JYA5BWQYSXDPRZJVLW3UGL3GJS",
	"FOPUBEUPBC6YLIQDLKL6EW775BMV7YOH",
	"GFK3RDAPQLLNCMQEVGGD2KCPZTLSG3HN",
	"H5EZTQE7ABFH27AUDTQFMZIALANK6RBG",
	"I2ADHGP4HL6J37NQAD73J7E5SKFIXJOT",
	"JEDZYC2HMGDBIDQKG3XSTXUSHMCBK725",
	//"MEJGDND55XNON7UU3ZKERJIZMMXJTVCV",   //we might not change more than 1 of the official witness
	//"MEJGDND55XNON7UU3ZKERJIZMMXJTVCV",
	//"MEJGDND55XNON7UU3ZKERJIZMMXJTVCV",
	"JPQKPRI5FMTQRJF4ZZMYZYDQVRD55OTC",
	"OYW2XTDKSNKGSEZ27LMGNOPJSYIXHBHC",
	"S7N5FE42F6ONPNDQLCF64E2MGFYKQR2I",
	"TKT4UESIKTTRALRRLWS4SENSTJX6ODCW",
	"UENJPVZ7HVHM6QGVGT6MWOJGGRTUTJXQ"
	];

assert.ok(arrWitnesses.length==12) //using the witness_list_unit: 'oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E=', (the genesis unit)
// arrWitnesses 是我自己的witness列表（轻节点自己可配置）

var arrFoundWitnesses=[]
var arrWitnessJoints=[]
var arrParentUnits=null
var arrLastBallUnits=[]
var assocLastBallByLastBallUnit={}

for (var i=0; i<unstable_mc_joints.length; i++){
	var objJoint = unstable_mc_joints[i];
	var objUnit = objJoint.unit;
	assert.equal(objectHash.getUnitHash(objUnit),objUnit.unit)
	if (arrParentUnits){
	  assert.ok(arrParentUnits.indexOf(objUnit.unit)>-1,arrParentUnits+" not found "+objUnit.unit)
	}
	var bAddedJoint = false;
	for (var j=0; j<objUnit.authors.length; j++){
		var address = objUnit.authors[j].address;
		if (arrWitnesses.indexOf(address) >= 0){
			if (arrFoundWitnesses.indexOf(address) === -1)
				arrFoundWitnesses.push(address);
			if (!bAddedJoint) {
				arrWitnessJoints.push(objJoint);
				console.log(i+", add witness authored unit: "+objJoint.unit.unit+", author is "+address);
			}
			bAddedJoint = true;
		}else{
			console.log(i+', no-witness authored unit:',objJoint.unit.unit," author is "+address);

		}
	}
	arrParentUnits = objUnit.parent_units;
	if (objUnit.last_ball_unit && arrFoundWitnesses.length >= constants.MAJORITY_OF_WITNESSES){
		arrLastBallUnits.push(objUnit.last_ball_unit);
		assocLastBallByLastBallUnit[objUnit.last_ball_unit] = objUnit.last_ball;
		console.log("                from unit:",objUnit.unit,", add last ball unit :",objUnit.last_ball_unit);
	}
}

//assert.ok( arrFoundWitnesses.length >= constants.MAJORITY_OF_WITNESSES, arrFoundWitnesses.length )
console.log("the unstable_mc_joins is",unstable_mc_joints.length,unstable_mc_joints.map(function(u){return u.unit.unit}))
console.log("the arrWitnessJoints is",arrWitnessJoints.length, " they are authored by witnesses", arrWitnessJoints.map(function(j){return j.unit.unit}))
console.log("the arrFoundWitnesses is",arrFoundWitnesses.length," > ",constants.MAJORITY_OF_WITNESSES, arrFoundWitnesses)
console.log("arrLastBallUnits",arrLastBallUnits.length,arrLastBallUnits)
console.log("assocLastBallByLastBallUnit",Object.keys(assocLastBallByLastBallUnit).length,'unit -> ball',assocLastBallByLastBallUnit)

var db = require('byteballcore/db.js');

function my_unit_sigature_verify_test(db) {
    var my_unit = joints[0].unit
    var my_unit_author_addr = my_unit.authors[0].address;
    var my_unit_author_sig = my_unit.authors[0].authentifiers.r
    var my_unit_author_def = null
	var my_unit_author_def_pubkey =null
	var signHash = objectHash.getUnitHashToSign(my_unit)
    var native_unit = objectHash.getNakedUnit(my_unit)
    for (var i=0; i<native_unit.authors.length; i++)
        delete native_unit.authors[i].authentifiers;
    console.log("before",util.inspect(my_unit,{depth:null}));
	console.log("after",native_unit);
    assert.ok(objectHash.getUnitHashToSign(native_unit).toString("hex")==signHash.toString("hex"))
    readDefinition(db, my_unit_author_addr, { //otherwise, lookup db for definition of the address
        ifFound: function (define) {
            my_unit_author_def = define;
            my_unit_author_def_pubkey=define[1].pubkey;
            console.log("my unit:", my_unit.unit, "author:", my_unit_author_addr, "signature:", my_unit_author_sig, "pubkey:",my_unit_author_def_pubkey)
            console.log("msg",signHash.toString("hex"),
                        "sig",(new Buffer(my_unit_author_sig,"base64")).toString("hex"),
                        "pubkey",(new Buffer(my_unit_author_def_pubkey,"base64")).toString("hex"));
            var res = ecdsaSig.verify(signHash, my_unit_author_sig, my_unit_author_def_pubkey);
            console.log("result=",res);

        },
        ifDefinitionNotFound: function (d) {
            console.log("definition not found for address " + my_unit_author_addr);
        }
    });
}

my_unit_sigature_verify_test(db)



// changes and definitions of witnesses
for (var i=0; i<witness_change_and_definition_joints.length; i++){
	var objJoint = witness_change_and_definition_joints[i];
	var objUnit = objJoint.unit;
	assert.ok(objJoint.ball);
	assert.ok(objectHash.getUnitHash(objUnit)==objUnit.unit);
	console.log("witness_change_and_definition_joints unit:",objUnit.unit)
	var bAuthoredByWitness = false;
	for (var j=0; j<objUnit.authors.length; j++){
		var address = objUnit.authors[j].address;
		if (arrWitnesses.indexOf(address) >= 0)
			bAuthoredByWitness = true;   //12个里面只要有一个对的上就是true，这个似乎有点问题！！！
	}
	assert.ok(bAuthoredByWitness==true)
}

objJoint=null
//console.log(objJoint)



var assocDefinitions = {};       // keyed by definition (aka the pubkey)  address -> def (pubkey)
var assocDefinitionChashes = {}; // keyed by address (hash160 of pubkey)  address -> address


//console.log("assocDefinitionChashes",assocDefinitionChashes)


function readDefinition(conn, definition_chash, callbacks){
    conn.query("SELECT definition FROM definitions WHERE definition_chash=?", [definition_chash], function(rows){
        if (rows.length === 0)
            return callbacks.ifDefinitionNotFound(definition_chash);
        callbacks.ifFound(JSON.parse(rows[0].definition));
    });
}

function readDefinitionByAddress(conn, address, max_mci, callbacks){
	var MAX_INT32 = Math.pow(2, 31) - 1;
	if (max_mci === null)
		max_mci = MAX_INT32;
	// try to find last definition change, otherwise definition_chash=address
	conn.query(
		"SELECT definition_chash FROM address_definition_changes CROSS JOIN units USING(unit) \n\
		WHERE address=? AND is_stable=1 AND sequence='good' AND main_chain_index<=? ORDER BY level DESC LIMIT 1",
		[address, max_mci],
		function(rows){
			var definition_chash = (rows.length > 0) ? rows[0].definition_chash : address;
			console.log("definition_chash is addr:",definition_chash)
			readDefinitionAtMci(conn, definition_chash, max_mci, callbacks);
		}
	);
}

// max_mci must be stable
function readDefinitionAtMci(conn, definition_chash, max_mci, callbacks){
	var sql = "SELECT definition FROM definitions CROSS JOIN unit_authors USING(definition_chash) CROSS JOIN units USING(unit) \n\
		WHERE definition_chash=? AND is_stable=1 AND sequence='good' AND main_chain_index<=?";
	var params = [definition_chash, max_mci];
	conn.query(sql, params, function(rows){
		if (rows.length === 0)
			return callbacks.ifDefinitionNotFound(definition_chash);
		callbacks.ifFound(JSON.parse(rows[0].definition));
	});
}

function validateAuthorSignaturesWithoutReferences(objAuthor, objUnit, arrAddressDefinition, callback){
    var objValidationState = {
        unit_hash_to_sign: objectHash.getUnitHashToSign(objUnit),
        last_ball_mci: -1,
        bNoReferences: true
    };
    Definition.validateAuthentifiers(
        null, objAuthor.address, null, arrAddressDefinition, objUnit, objValidationState, objAuthor.authentifiers,
        function(err, res){
            if (err) // error in address definition
                return callback(err);
            if (!res) // wrong signature or the like
                return callback("authentifier verification failed");
            callback();
        }
    );
}


function validateUnit(objUnit, bRequireDefinitionOrChange, cb2){
	console.log("validateUnit",objUnit.unit, "bRequireDefinitionOrChange="+bRequireDefinitionOrChange);
    var bFound = false;
    async.eachSeries(
        objUnit.authors,
        function(author, cb3){
        	console.log("--","validate","author",author.address)
            var address = author.address;
        	console.log("----","if",address,"in arrWitnesses",arrWitnesses.indexOf(address)>=0)
            if (arrWitnesses.indexOf(address) === -1) // not a witness - skip it
                return cb3();
            var definition_chash = assocDefinitionChashes[address];
            console.log("----","get definition hash (address),",definition_chash)
            if (!definition_chash)
                throw Error("definition chash not known for address "+address);
            if (author.definition){
                console.log("----","author has definition(pubkey)",JSON.stringify(author))
                //console.log("              ","author ",author.address,"pubkey",author.definition)
                if (objectHash.getChash160(author.definition) !== definition_chash)  //address is pubkey's hash160
                    return cb3("definition doesn't hash to the expected value");
                assocDefinitions[definition_chash] = author.definition;
                bFound = true;
            }else{
                console.log("----","author has no definition (no pubkey)",JSON.stringify(author))

			}

            function handleAuthor(){
                // FIX
                validateAuthorSignaturesWithoutReferences(author, objUnit, assocDefinitions[definition_chash], function(err){
                    if (err)
                        return cb3(err);
                    for (var i=0; i<objUnit.messages.length; i++){
                        var message = objUnit.messages[i];
                        if (message.app === 'address_definition_change'
                            && (message.payload.address === address || objUnit.authors.length === 1 && objUnit.authors[0].address === address)){
                            assocDefinitionChashes[address] = message.payload.definition_chash;
                            bFound = true;
                        }
                    }
                    cb3();
                });
            }

            //console.log("--",assocDefinitions)
            if (assocDefinitions[definition_chash])
                return handleAuthor();  //since the unit is created by witness which always has definition, go directly here
            readDefinition(db, definition_chash, { //otherwise, lookup db for definition of the address
                ifFound: function(arrDefinition){
                    assocDefinitions[definition_chash] = arrDefinition;
                    handleAuthor();
                },
                ifDefinitionNotFound: function(d){
                    throw Error("definition "+definition_chash+" not found, address "+address);
                }
            });
        },
        function(err){
            if (err)
                return cb2(err);
            if (bRequireDefinitionOrChange && !bFound)
                return cb2("neither definition nor change");
            cb2();
        }
    ); // each authors
}

var async = require('async');
var Definition = require("./my_def.js");
var bFromCurrent = false

function processWitnessProof(handleResult){
    async.series([
        function (cb) { // read latest known definitions of witness addresses
            console.log("1. read latest known definitions of witness addresses")
            if (!bFromCurrent) {
                arrWitnesses.forEach(function (address) {
                    assocDefinitionChashes[address] = address;
                });
                return cb(); //go to step 2 directly
            }
            async.eachSeries(
                arrWitnesses,
                function (address, cb2) {
                    storage.readDefinitionByAddress(db, address, null, {
                        ifFound: function (arrDefinition) {
                            var definition_chash = objectHash.getChash160(arrDefinition);
                            assocDefinitions[definition_chash] = arrDefinition;
                            assocDefinitionChashes[address] = definition_chash;
                            cb2();
                        },
                        ifDefinitionNotFound: function (definition_chash) {
                            assocDefinitionChashes[address] = definition_chash;
                            cb2();
                        }
                    });
                },
                cb
            );
        },
        function (cb) { // handle changes of definitions
            console.log("2. handle changes of definitions");
            async.eachSeries(
                witness_change_and_definition_joints,
                function (objJoint, cb2) {
                    var objUnit = objJoint.unit;
                    if (!bFromCurrent)
                        return validateUnit(objUnit, true, cb2);  //return to step 3 directly
                    db.query("SELECT 1 FROM units WHERE unit=? AND is_stable=1", [objUnit.unit], function (rows) {
                        if (rows.length > 0) // already known and stable - skip it
                            return cb2();
                        validateUnit(objUnit, true, cb2);
                    });
                },
                cb
            ); // each change or definition
        },
        function (cb) { // check signatures of unstable witness joints
            console.log("3. check signatures of unstable witness jointss")
            async.eachSeries(
                arrWitnessJoints.reverse(), // they came in reverse chronological order, reverse() reverses in place
                function (objJoint, cb2) {
                    validateUnit(objJoint.unit, false, cb2);
                },
                cb
            );
        },
    ], function (err) {
        console.log("error:", err)
        err ? handleResult(err) : handleResult(null, arrLastBallUnits, assocLastBallByLastBallUnit);
    });
}


/*
processWitnessProof(
    function (err, arrLastBallUnits, assocLastBallByLastBallUnit) {
        console.log(err, arrLastBallUnits, assocLastBallByLastBallUnit);
    }
);
*/






