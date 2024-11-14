-- sqlfmt d:postgres

--
-- PostgreSQL database dump
--

-- Dumped from database version 14.13
-- Dumped by pg_dump version 14.13

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: rt_nutrient; Type: TABLE; Schema: fdc; Owner: app_owner
--

CREATE TABLE fdc.rt_nutrient (
    id smallint NOT NULL,
    unit_id smallint,
    use_in_summary boolean DEFAULT true,
    name text,
    common_name text,
    remarks text
);


ALTER TABLE fdc.rt_nutrient OWNER TO app_owner;

--
-- Name: rt_unit; Type: TABLE; Schema: fdc; Owner: app_owner
--

CREATE TABLE fdc.rt_unit (
    id smallint NOT NULL,
    name text NOT NULL,
    label text,
    remarks text
);


ALTER TABLE fdc.rt_unit OWNER TO app_owner;

--
-- Name: rt_unit_id_seq; Type: SEQUENCE; Schema: fdc; Owner: app_owner
--

CREATE SEQUENCE fdc.rt_unit_id_seq
    AS smallint
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE fdc.rt_unit_id_seq OWNER TO app_owner;

--
-- Name: rt_unit_id_seq; Type: SEQUENCE OWNED BY; Schema: fdc; Owner: app_owner
--

ALTER SEQUENCE fdc.rt_unit_id_seq OWNED BY fdc.rt_unit.id;


--
-- Name: rt_unit id; Type: DEFAULT; Schema: fdc; Owner: app_owner
--

ALTER TABLE ONLY fdc.rt_unit ALTER COLUMN id SET DEFAULT nextval('fdc.rt_unit_id_seq'::regclass);


--
-- Data for Name: rt_nutrient; Type: TABLE DATA; Schema: fdc; Owner: app_owner
--

COPY fdc.rt_nutrient (id, unit_id, use_in_summary, name, common_name, remarks) FROM stdin;
2047	4	t	Energy (Atwater General Factors)	\N	\N
2048	4	t	Energy (Atwater Specific Factors)	\N	\N
1001	10	t	Solids	\N	\N
1002	10	t	Nitrogen	\N	\N
1003	10	t	Protein	\N	\N
1004	10	t	Total lipid (fat)	\N	\N
1005	10	t	Carbohydrate, by difference	\N	\N
1007	10	t	Ash	\N	\N
1008	4	t	Energy	\N	\N
1009	10	t	Starch	\N	\N
1010	10	t	Sucrose	\N	\N
1011	10	t	Glucose	\N	\N
1012	10	t	Fructose	\N	\N
1013	10	t	Lactose	\N	\N
1014	10	t	Maltose	\N	\N
1015	10	t	Amylose	\N	\N
1016	10	t	Amylopectin	\N	\N
1017	10	t	Pectin	\N	\N
1018	10	t	Alcohol, ethyl	\N	\N
1019	10	t	Pentosan	\N	\N
1020	10	t	Pentoses	\N	\N
1021	10	t	Hemicellulose	\N	\N
1022	10	t	Cellulose	\N	\N
1023	6	t	pH	\N	\N
1024	5	t	Specific Gravity	\N	\N
1025	10	t	Organic acids	\N	\N
1026	11	t	Acetic acid	\N	\N
1027	11	t	Aconitic acid	\N	\N
1028	11	t	Benzoic acid	\N	\N
1029	11	t	Chelidonic acid	\N	\N
1030	11	t	Chlorogenic acid	\N	\N
1031	11	t	Cinnamic acid	\N	\N
1032	11	t	Citric acid	\N	\N
1033	11	t	Fumaric acid	\N	\N
1034	11	t	Galacturonic acid	\N	\N
1035	11	t	Gallic acid	\N	\N
1036	11	t	Glycolic acid	\N	\N
1037	11	t	Isocitric acid	\N	\N
1038	11	t	Lactic acid	\N	\N
1039	11	t	Malic acid	\N	\N
1040	11	t	Oxaloacetic acid	\N	\N
1041	11	t	Oxalic acid	\N	\N
1042	11	t	Phytic acid	\N	\N
1043	11	t	Pyruvic acid	\N	\N
1044	11	t	Quinic acid	\N	\N
1045	11	t	Salicylic acid	\N	\N
1046	11	t	Succinic acid	\N	\N
1047	11	t	Tartaric acid	\N	\N
1048	11	t	Ursolic acid	\N	\N
1049	10	t	Solids, non-fat	\N	\N
1050	10	t	Carbohydrate, by summation	\N	\N
1051	10	t	Water	\N	\N
1052	10	t	Adjusted Nitrogen	\N	\N
1053	10	t	Adjusted Protein	\N	\N
1054	10	t	Piperine	\N	\N
1055	10	t	Mannitol	\N	\N
1056	10	t	Sorbitol	\N	\N
1057	11	t	Caffeine	\N	\N
1058	11	t	Theobromine	\N	\N
1059	11	t	Nitrates	\N	\N
1060	11	t	Nitrites	\N	\N
1061	11	t	Nitrosamine,total	\N	\N
1062	2	t	Energy	\N	\N
1063	10	t	Sugars, Total	\N	\N
1064	10	t	Solids, soluble	\N	\N
1065	10	t	Glycogen	\N	\N
1067	10	t	Reducing sugars	\N	\N
1068	10	t	Beta-glucans	\N	\N
1069	10	t	Oligosaccharides	\N	\N
1070	10	t	Nonstarch polysaccharides	\N	\N
1071	10	t	Resistant starch	\N	\N
1072	10	t	Carbohydrate, other	\N	\N
1073	10	t	Arabinose	\N	\N
1074	10	t	Xylose	\N	\N
1075	10	t	Galactose	\N	\N
1076	10	t	Raffinose	\N	\N
1077	10	t	Stachyose	\N	\N
1078	10	t	Xylitol	\N	\N
1079	10	t	Fiber, total dietary	\N	\N
1080	10	t	Lignin	\N	\N
1081	10	t	Ribose	\N	\N
1082	10	t	Fiber, soluble	\N	\N
1083	11	t	Theophylline	\N	\N
1084	10	t	Fiber, insoluble	\N	\N
1085	10	t	Total fat (NLEA)	\N	\N
1086	10	t	Total sugar alcohols	\N	\N
1087	11	t	Calcium, Ca	\N	\N
1088	11	t	Chlorine, Cl	\N	\N
1089	11	t	Iron, Fe	\N	\N
1090	11	t	Magnesium, Mg	\N	\N
1091	11	t	Phosphorus, P	\N	\N
1092	11	t	Potassium, K	\N	\N
1093	11	t	Sodium, Na	\N	\N
1094	11	t	Sulfur, S	\N	\N
1095	11	t	Zinc, Zn	\N	\N
1096	7	t	Chromium, Cr	\N	\N
1097	7	t	Cobalt, Co	\N	\N
1098	11	t	Copper, Cu	\N	\N
1099	7	t	Fluoride, F	\N	\N
1100	7	t	Iodine, I	\N	\N
1101	11	t	Manganese, Mn	\N	\N
1102	7	t	Molybdenum, Mo	\N	\N
1103	7	t	Selenium, Se	\N	\N
1104	12	t	Vitamin A, IU	\N	\N
1105	7	t	Retinol	\N	\N
1106	7	t	Vitamin A, RAE	\N	\N
1107	7	t	Carotene, beta	\N	\N
1108	7	t	Carotene, alpha	\N	\N
1109	11	t	Vitamin E (alpha-tocopherol)	\N	\N
1110	12	t	Vitamin D (D2 + D3), International Units	\N	\N
1111	7	t	Vitamin D2 (ergocalciferol)	\N	\N
1112	7	t	Vitamin D3 (cholecalciferol)	\N	\N
1113	7	t	25-hydroxycholecalciferol	\N	\N
1114	7	t	Vitamin D (D2 + D3)	\N	\N
1115	7	t	25-hydroxyergocalciferol	\N	\N
1116	7	t	Phytoene	\N	\N
1117	7	t	Phytofluene	\N	\N
1118	7	t	Carotene, gamma	\N	\N
1119	7	t	Zeaxanthin	\N	\N
1120	7	t	Cryptoxanthin, beta	\N	\N
1121	7	t	Lutein	\N	\N
1122	7	t	Lycopene	\N	\N
1123	7	t	Lutein + zeaxanthin	\N	\N
1124	12	t	Vitamin E (label entry primarily)	\N	\N
1125	11	t	Tocopherol, beta	\N	\N
1126	11	t	Tocopherol, gamma	\N	\N
1127	11	t	Tocopherol, delta	\N	\N
1128	11	t	Tocotrienol, alpha	\N	\N
1129	11	t	Tocotrienol, beta	\N	\N
1130	11	t	Tocotrienol, gamma	\N	\N
1131	11	t	Tocotrienol, delta	\N	\N
1132	7	t	Aluminum, Al	\N	\N
1133	7	t	Antimony, Sb	\N	\N
1134	7	t	Arsenic, As	\N	\N
1135	7	t	Barium, Ba	\N	\N
1136	7	t	Beryllium, Be	\N	\N
1137	7	t	Boron, B	\N	\N
1138	7	t	Bromine, Br	\N	\N
1139	7	t	Cadmium, Cd	\N	\N
1140	7	t	Gold, Au	\N	\N
1141	11	t	Iron, heme	\N	\N
1142	11	t	Iron, non-heme	\N	\N
1143	7	t	Lead, Pb	\N	\N
1144	7	t	Lithium, Li	\N	\N
1145	7	t	Mercury, Hg	\N	\N
1146	7	t	Nickel, Ni	\N	\N
1147	7	t	Rubidium, Rb	\N	\N
1149	11	t	Salt, NaCl	\N	\N
1150	7	t	Silicon, Si	\N	\N
1151	7	t	Silver, Ag	\N	\N
1152	7	t	Strontium, Sr	\N	\N
1153	7	t	Tin, Sn	\N	\N
1154	7	t	Titanium, Ti	\N	\N
1155	7	t	Vanadium, V	\N	\N
1156	3	t	Vitamin A, RE	\N	\N
1157	3	t	Carotene	\N	\N
1158	1	t	Vitamin E	\N	\N
1159	7	t	cis-beta-Carotene	\N	\N
1160	7	t	cis-Lycopene	\N	\N
1161	7	t	cis-Lutein/Zeaxanthin	\N	\N
1162	11	t	Vitamin C, total ascorbic acid	\N	\N
1163	11	t	Vitamin C, reduced ascorbic acid	\N	\N
1164	11	t	Vitamin C, dehydro ascorbic acid	\N	\N
1165	11	t	Thiamin	\N	\N
1166	11	t	Riboflavin	\N	\N
1167	11	t	Niacin	\N	\N
1168	11	t	Niacin from tryptophan, determined	\N	\N
1169	11	t	Niacin equivalent N406 +N407	\N	\N
1170	11	t	Pantothenic acid	\N	\N
1171	11	t	Vitamin B-6, pyridoxine, alcohol form	\N	\N
1172	11	t	Vitamin B-6, pyridoxal, aldehyde form	\N	\N
1173	11	t	Vitamin B-6, pyridoxamine, amine form	\N	\N
1174	11	t	Vitamin B-6, N411 + N412 +N413	\N	\N
1175	11	t	Vitamin B-6	\N	\N
1176	7	t	Biotin	\N	\N
1177	7	t	Folate, total	\N	\N
1178	7	t	Vitamin B-12	\N	\N
1179	7	t	Folate, free	\N	\N
1180	11	t	Choline, total	\N	\N
1181	11	t	Inositol	\N	\N
1182	11	t	Inositol phosphate	\N	\N
1183	7	t	Vitamin K (Menaquinone-4)	\N	\N
1184	7	t	Vitamin K (Dihydrophylloquinone)	\N	\N
1185	7	t	Vitamin K (phylloquinone)	\N	\N
1186	7	t	Folic acid	\N	\N
1187	7	t	Folate, food	\N	\N
1188	7	t	5-methyl tetrahydrofolate (5-MTHF)	\N	\N
1189	7	t	Folate, not 5-MTHF	\N	\N
1190	7	t	Folate, DFE	\N	\N
1191	7	t	10-Formyl folic acid (10HCOFA)	\N	\N
1192	7	t	5-Formyltetrahydrofolic acid (5-HCOH4	\N	\N
1193	7	t	Tetrahydrofolic acid (THF)	\N	\N
1194	11	t	Choline, free	\N	\N
1195	11	t	Choline, from phosphocholine	\N	\N
1196	11	t	Choline, from phosphotidyl choline	\N	\N
1197	11	t	Choline, from glycerophosphocholine	\N	\N
1198	11	t	Betaine	\N	\N
1199	11	t	Choline, from sphingomyelin	\N	\N
1200	11	t	p-Hydroxy benzoic acid	\N	\N
1201	11	t	Caffeic acid	\N	\N
1202	11	t	p-Coumaric acid	\N	\N
1203	11	t	Ellagic acid	\N	\N
1204	11	t	Ferrulic acid	\N	\N
1205	11	t	Gentisic acid	\N	\N
1206	11	t	Tyrosol	\N	\N
1207	11	t	Vanillic acid	\N	\N
1208	11	t	Phenolic acids, total	\N	\N
1209	11	t	Polyphenols, total	\N	\N
1210	10	t	Tryptophan	\N	\N
1211	10	t	Threonine	\N	\N
1212	10	t	Isoleucine	\N	\N
1213	10	t	Leucine	\N	\N
1214	10	t	Lysine	\N	\N
1215	10	t	Methionine	\N	\N
1216	10	t	Cystine	\N	\N
1217	10	t	Phenylalanine	\N	\N
1218	10	t	Tyrosine	\N	\N
1219	10	t	Valine	\N	\N
1220	10	t	Arginine	\N	\N
1221	10	t	Histidine	\N	\N
1222	10	t	Alanine	\N	\N
1223	10	t	Aspartic acid	\N	\N
1224	10	t	Glutamic acid	\N	\N
1225	10	t	Glycine	\N	\N
1226	10	t	Proline	\N	\N
1227	10	t	Serine	\N	\N
1228	10	t	Hydroxyproline	\N	\N
1229	10	t	Cysteine and methionine(sulfer containig AA)	\N	\N
1230	10	t	Phenylalanine and tyrosine (aromatic  AA)	\N	\N
1231	10	t	Asparagine	\N	\N
1232	10	t	Cysteine	\N	\N
1233	10	t	Glutamine	\N	\N
1234	10	t	Taurine	\N	\N
1235	10	t	Sugars, added	\N	\N
1236	10	t	Sugars, intrinsic	\N	\N
1237	11	t	Calcium, added	\N	\N
1238	11	t	Iron, added	\N	\N
1239	11	t	Calcium, intrinsic	\N	\N
1240	11	t	Iron, intrinsic	\N	\N
1241	11	t	Vitamin C, added	\N	\N
1242	11	t	Vitamin E, added	\N	\N
1243	11	t	Thiamin, added	\N	\N
1244	11	t	Riboflavin, added	\N	\N
1245	11	t	Niacin, added	\N	\N
1246	7	t	Vitamin B-12, added	\N	\N
1247	11	t	Vitamin C, intrinsic	\N	\N
1248	11	t	Vitamin E, intrinsic	\N	\N
1249	11	t	Thiamin, intrinsic	\N	\N
1250	11	t	Riboflavin, intrinsic	\N	\N
1251	11	t	Niacin, intrinsic	\N	\N
1252	7	t	Vitamin B-12, intrinsic	\N	\N
1253	11	t	Cholesterol	\N	\N
1254	10	t	Glycerides	\N	\N
1255	10	t	Phospholipids	\N	\N
1256	10	t	Glycolipids	\N	\N
1257	10	t	Fatty acids, total trans	\N	\N
1258	10	t	Fatty acids, total saturated	\N	\N
1259	10	t	SFA 4:0	\N	\N
1260	10	t	SFA 6:0	\N	\N
1261	10	t	SFA 8:0	\N	\N
1262	10	t	SFA 10:0	\N	\N
1263	10	t	SFA 12:0	\N	\N
1264	10	t	SFA 14:0	\N	\N
1265	10	t	SFA 16:0	\N	\N
1266	10	t	SFA 18:0	\N	\N
1267	10	t	SFA 20:0	\N	\N
1268	10	t	MUFA 18:1	\N	\N
1269	10	t	PUFA 18:2	\N	\N
1270	10	t	PUFA 18:3	\N	\N
1271	10	t	PUFA 20:4	\N	\N
1272	10	t	PUFA 22:6 n-3 (DHA)	\N	\N
1273	10	t	SFA 22:0	\N	\N
1274	10	t	MUFA 14:1	\N	\N
1275	10	t	MUFA 16:1	\N	\N
1276	10	t	PUFA 18:4	\N	\N
1277	10	t	MUFA 20:1	\N	\N
1278	10	t	PUFA 20:5 n-3 (EPA)	\N	\N
1279	10	t	MUFA 22:1	\N	\N
1280	10	t	PUFA 22:5 n-3 (DPA)	\N	\N
1281	10	t	TFA 14:1 t	\N	\N
1283	11	t	Phytosterols	\N	\N
1284	11	t	Ergosterol	\N	\N
1285	11	t	Stigmasterol	\N	\N
1286	11	t	Campesterol	\N	\N
1287	11	t	Brassicasterol	\N	\N
1288	11	t	Beta-sitosterol	\N	\N
1289	11	t	Campestanol	\N	\N
1290	10	t	Unsaponifiable matter (lipids)	\N	\N
1291	10	t	Fatty acids, other than 607-615, 617-621, 624-632, 652-654, 686-689)	\N	\N
1292	10	t	Fatty acids, total monounsaturated	\N	\N
1293	10	t	Fatty acids, total polyunsaturated	\N	\N
1294	11	t	Beta-sitostanol	\N	\N
1295	11	t	Delta-7-avenasterol	\N	\N
1296	11	t	Delta-5-avenasterol	\N	\N
1297	11	t	Alpha-spinasterol	\N	\N
1298	11	t	Phytosterols, other	\N	\N
1299	10	t	SFA 15:0	\N	\N
1300	10	t	SFA 17:0	\N	\N
1301	10	t	SFA 24:0	\N	\N
1302	10	t	Wax Esters(Total Wax)	\N	\N
1303	10	t	TFA 16:1 t	\N	\N
1304	10	t	TFA 18:1 t	\N	\N
1305	10	t	TFA 22:1 t	\N	\N
1306	10	t	TFA 18:2 t not further defined	\N	\N
1307	10	t	PUFA 18:2 i	\N	\N
1308	10	t	PUFA 18:2 t,c	\N	\N
1309	10	t	PUFA 18:2 c,t	\N	\N
1310	10	t	TFA 18:2 t,t	\N	\N
1311	10	t	PUFA 18:2 CLAs	\N	\N
1312	10	t	MUFA 24:1 c	\N	\N
1313	10	t	PUFA 20:2 n-6 c,c	\N	\N
1314	10	t	MUFA 16:1 c	\N	\N
1315	10	t	MUFA 18:1 c	\N	\N
1316	10	t	PUFA 18:2 n-6 c,c	\N	\N
1317	10	t	MUFA 22:1 c	\N	\N
1318	10	t	Fatty acids, saturated, other	\N	\N
1319	10	t	Fatty acids, monounsat., other	\N	\N
1320	10	t	Fatty acids, polyunsat., other	\N	\N
1321	10	t	PUFA 18:3 n-6 c,c,c	\N	\N
1322	10	t	SFA 19:0	\N	\N
1323	10	t	MUFA 17:1	\N	\N
1324	10	t	PUFA 16:2	\N	\N
1325	10	t	PUFA 20:3	\N	\N
1326	10	t	Fatty acids, total sat., NLEA	\N	\N
1327	10	t	Fatty acids, total monounsat., NLEA	\N	\N
1328	10	t	Fatty acids, total polyunsat., NLEA	\N	\N
1329	10	t	Fatty acids, total trans-monoenoic	\N	\N
1330	10	t	Fatty acids, total trans-dienoic	\N	\N
1331	10	t	Fatty acids, total trans-polyenoic	\N	\N
1332	10	t	SFA 13:0	\N	\N
1333	10	t	MUFA 15:1	\N	\N
1334	10	t	PUFA 22:2	\N	\N
1335	10	t	SFA 11:0	\N	\N
1336	9	t	ORAC, Hydrophyllic	\N	\N
1337	9	t	ORAC, Lipophillic	\N	\N
1338	9	t	ORAC, Total	\N	\N
1339	8	t	Total Phenolics	\N	\N
1340	11	t	Daidzein	\N	\N
1341	11	t	Genistein	\N	\N
1342	11	t	Glycitein	\N	\N
1343	11	t	Isoflavones	\N	\N
1344	11	t	Biochanin A	\N	\N
1345	11	t	Formononetin	\N	\N
1346	11	t	Coumestrol	\N	\N
1347	11	t	Flavonoids, total	\N	\N
1348	11	t	Anthocyanidins	\N	\N
1349	11	t	Cyanidin	\N	\N
1350	11	t	Proanthocyanidin (dimer-A linkage)	\N	\N
1351	11	t	Proanthocyanidin monomers	\N	\N
1352	11	t	Proanthocyanidin dimers	\N	\N
1353	11	t	Proanthocyanidin trimers	\N	\N
1354	11	t	Proanthocyanidin 4-6mers	\N	\N
1355	11	t	Proanthocyanidin 7-10mers	\N	\N
1356	11	t	Proanthocyanidin polymers (>10mers)	\N	\N
1357	11	t	Delphinidin	\N	\N
1358	11	t	Malvidin	\N	\N
1359	11	t	Pelargonidin	\N	\N
1360	11	t	Peonidin	\N	\N
1361	11	t	Petunidin	\N	\N
1362	11	t	Flavans, total	\N	\N
1363	11	t	Catechins, total	\N	\N
1364	11	t	Catechin	\N	\N
1365	11	t	Epigallocatechin	\N	\N
1366	11	t	Epicatechin	\N	\N
1367	11	t	Epicatechin-3-gallate	\N	\N
1368	11	t	Epigallocatechin-3-gallate	\N	\N
1369	11	t	Procyanidins, total	\N	\N
1370	11	t	Theaflavins	\N	\N
1371	11	t	Thearubigins	\N	\N
1372	11	t	Flavanones, total	\N	\N
1373	11	t	Eriodictyol	\N	\N
1374	11	t	Hesperetin	\N	\N
1375	11	t	Isosakuranetin	\N	\N
1376	11	t	Liquiritigenin	\N	\N
1377	11	t	Naringenin	\N	\N
1378	11	t	Flavones, total	\N	\N
1379	11	t	Apigenin	\N	\N
1380	11	t	Chrysoeriol	\N	\N
1381	11	t	Diosmetin	\N	\N
1382	11	t	Luteolin	\N	\N
1383	11	t	Nobiletin	\N	\N
1384	11	t	Sinensetin	\N	\N
1385	11	t	Tangeretin	\N	\N
1386	11	t	Flavonols, total	\N	\N
1387	11	t	Isorhamnetin	\N	\N
1388	11	t	Kaempferol	\N	\N
1389	11	t	Limocitrin	\N	\N
1390	11	t	Myricetin	\N	\N
1391	11	t	Quercetin	\N	\N
1392	11	t	Theogallin	\N	\N
1393	11	t	Theaflavin -3,3' -digallate	\N	\N
1394	11	t	Theaflavin -3' -gallate	\N	\N
1395	11	t	Theaflavin -3 -gallate	\N	\N
1396	11	t	(+) -Gallo catechin	\N	\N
1397	11	t	(+)-Catechin 3-gallate	\N	\N
1398	11	t	(+)-Gallocatechin 3-gallate	\N	\N
1399	10	t	Mannose	\N	\N
1400	10	t	Triose	\N	\N
1401	10	t	Tetrose	\N	\N
1402	10	t	Other Saccharides	\N	\N
1403	10	t	Inulin	\N	\N
1404	10	t	PUFA 18:3 n-3 c,c,c (ALA)	\N	\N
1405	10	t	PUFA 20:3 n-3	\N	\N
1406	10	t	PUFA 20:3 n-6	\N	\N
1407	10	t	PUFA 20:4 n-3	\N	\N
1408	10	t	PUFA 20:4 n-6	\N	\N
1409	10	t	PUFA 18:3i	\N	\N
1410	10	t	PUFA 21:5	\N	\N
1411	10	t	PUFA 22:4	\N	\N
1412	10	t	MUFA 18:1-11 t (18:1t n-7)	\N	\N
1413	10	t	MUFA 18:1-11 c (18:1c n-7)	\N	\N
1414	10	t	PUFA 20:3 n-9	\N	\N
2000	10	t	Sugars, total including NLEA	\N	\N
2003	10	t	SFA 5:0	\N	\N
2004	10	t	SFA 7:0	\N	\N
2005	10	t	SFA 9:0	\N	\N
2006	10	t	SFA 21:0	\N	\N
2007	10	t	SFA 23:0	\N	\N
2008	10	t	MUFA 12:1	\N	\N
2009	10	t	MUFA 14:1 c	\N	\N
2010	10	t	MUFA 17:1 c	\N	\N
2011	10	t	TFA 17:1 t	\N	\N
2012	10	t	MUFA 20:1 c	\N	\N
2013	10	t	TFA 20:1 t	\N	\N
2014	10	t	MUFA 22:1 n-9	\N	\N
2015	10	t	MUFA 22:1 n-11	\N	\N
2016	10	t	PUFA 18:2 c	\N	\N
2017	10	t	TFA 18:2 t	\N	\N
2018	10	t	PUFA 18:3 c	\N	\N
2019	10	t	TFA 18:3 t	\N	\N
2020	10	t	PUFA 20:3 c	\N	\N
2021	10	t	PUFA 22:3	\N	\N
2022	10	t	PUFA 20:4c	\N	\N
2023	10	t	PUFA 20:5c	\N	\N
2024	10	t	PUFA 22:5 c	\N	\N
2025	10	t	PUFA 22:6 c	\N	\N
2026	10	t	PUFA 20:2 c	\N	\N
2027	10	t	Proximate	\N	\N
2028	7	t	trans-beta-Carotene	\N	\N
2029	7	t	trans-Lycopene	\N	\N
2032	7	t	Cryptoxanthin, alpha	\N	\N
2033	10	t	Total dietary fiber (AOAC 2011.25)	\N	\N
2034	10	t	Insoluble dietary fiber (IDF)	\N	\N
2035	10	t	Soluble dietary fiber (SDFP+SDFS)	\N	\N
2036	10	t	Soluble dietary fiber (SDFP)	\N	\N
2037	10	t	Soluble dietary fiber (SDFS)	\N	\N
2038	10	t	High Molecular Weight Dietary Fiber (HMWDF)	\N	\N
2039	10	t	Carbohydrates	\N	\N
2040	7	t	Other carotenoids	\N	\N
2041	11	t	Tocopherols and tocotrienols	\N	\N
2042	10	t	Amino acids	\N	\N
2043	11	t	Minerals	\N	\N
2044	10	t	Lipids	\N	\N
2045	10	t	Proximates	\N	\N
2046	10	t	Vitamins and Other Components	\N	\N
2055	11	t	Total Tocopherols	\N	\N
2054	11	t	Total Tocotrienols	\N	\N
2053	11	t	Stigmastadiene	\N	\N
2052	11	t	Delta-7-Stigmastenol	\N	\N
2049	11	t	Daidzin	\N	\N
2050	11	t	Genistin	\N	\N
2051	11	t	Glycitin	\N	\N
2057	11	t	Ergothioneine	\N	\N
2058	10	t	Beta-glucan	\N	\N
2059	7	t	Vitamin D4	\N	\N
2060	11	t	Ergosta-7-enol	\N	\N
2061	11	t	 Ergosta-7,22-dienol	\N	\N
2062	11	t	 Ergosta-5,7-dienol	\N	\N
2063	10	t	Verbascose	\N	\N
2064	11	t	Oligosaccharides	\N	\N
2065	10	t	Low Molecular Weight Dietary Fiber (LMWDF)	\N	\N
-1	10	t	MUFA	\N	\N
-2	10	t	PUFA	\N	\N
-3	10	t	SFA	\N	\N
-4	10	t	TFA	\N	\N
\.


--
-- Data for Name: rt_unit; Type: TABLE DATA; Schema: fdc; Owner: app_owner
--

COPY fdc.rt_unit (id, name, label, remarks) FROM stdin;
7	UG	μg	micro-grams
9	UMOL_TE	μMol TE	micro-moles of trolox equivalents
5	SP_GR	sp gr	Specific gravity?
6	PH	pH	pH
8	MG_GAE	mg GAE	milli-grams of gallic acid equivalents
1	MG_ATE	mg ATE	milli-grams of alpha tocopherol equivalent
11	MG	mg	milli-grams
3	MCG_RE	μg RE	micro-grams of retinol equivalent
2	kJ	kJ	kilo-Joules
4	KCAL	kCal	kilo-calories
12	IU	IU	International Unit is a measure of biological activity and is different for each substance
10	G	g	grams
\.


--
-- Name: rt_unit_id_seq; Type: SEQUENCE SET; Schema: fdc; Owner: app_owner
--

SELECT pg_catalog.setval('fdc.rt_unit_id_seq', 12, true);


--
-- Name: rt_nutrient rt_nutrient_nk; Type: CONSTRAINT; Schema: fdc; Owner: app_owner
--

ALTER TABLE ONLY fdc.rt_nutrient
    ADD CONSTRAINT rt_nutrient_nk UNIQUE (unit_id, name);


--
-- Name: rt_nutrient rt_nutrient_pk; Type: CONSTRAINT; Schema: fdc; Owner: app_owner
--

ALTER TABLE ONLY fdc.rt_nutrient
    ADD CONSTRAINT rt_nutrient_pk PRIMARY KEY (id);


--
-- Name: rt_unit rt_unit_nk; Type: CONSTRAINT; Schema: fdc; Owner: app_owner
--

ALTER TABLE ONLY fdc.rt_unit
    ADD CONSTRAINT rt_unit_nk UNIQUE (name);


--
-- Name: rt_unit rt_unit_pk; Type: CONSTRAINT; Schema: fdc; Owner: app_owner
--

ALTER TABLE ONLY fdc.rt_unit
    ADD CONSTRAINT rt_unit_pk PRIMARY KEY (id);


--
-- Name: rt_nutrient rt_nutrient_fk01; Type: FK CONSTRAINT; Schema: fdc; Owner: app_owner
--

ALTER TABLE ONLY fdc.rt_nutrient
    ADD CONSTRAINT rt_nutrient_fk01 FOREIGN KEY (unit_id) REFERENCES fdc.rt_unit(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--
